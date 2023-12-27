package http

import (
	"fmt"
	"github.com/pygrum/monarch/pkg/types"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/transport"
)

const (
	StatusActive   SessionStatus = "active"
	StatusInactive SessionStatus = "inactive"
	StatusKilled   SessionStatus = "killed"
)

var (
	// new secret each restart
	key = []byte(uuid.New().String())
)

type SessionStatus string

type HTTPSession struct {
	ID            int
	RequestQueue  types.Queue
	ResponseQueue types.Queue
	Agent         *db.Agent
	LastActive    time.Time
	Status        SessionStatus
	lock          sync.Mutex
	Authenticated bool
	Killer        chan struct{}
	Info          transport.Registration
	SentRequests  map[string]int
	UsedBy        string
}

type Claims struct {
	jwt.RegisteredClaims
	ID int
}

type sessions struct {
	lock           sync.Mutex
	count          int
	sessionMap     map[int]*HTTPSession
	sortedSessions []*HTTPSession
}

// newSession registers a session and creates a cookie for auth
func (s *sessions) newSession(agent *db.Agent, isTCP bool, connectInfo *transport.Registration) (string, time.Time, int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	id := s.count
	// Don't check for existing session before connect anymore = people can spam the server with new sessions
	// To stop spamming, delete the agent that is being used.

	//// check if session with given agent exists anywhere
	//for i, sess := range s.sortedSessions {
	//	if agent.AgentID == sess.Agent.AgentID {
	//		if !isTCP {
	//			if !sess.Authenticated {
	//				return "", time.Time{}, 0, fmt.Errorf("agent %s was not previously authenticated", agent.AgentID)
	//			}
	//		}
	//		delete(s.sessionMap, sess.ID) // remove session from map
	//		s.sortedSessions = append(s.sortedSessions[:i], s.sortedSessions[i+1:]...)
	//	}
	//}
	newSession := &HTTPSession{
		ID:            id,
		RequestQueue:  NewRequestQueue(),
		ResponseQueue: NewResponseQueue(),
		Agent:         agent,
		Killer:        make(chan struct{}),
		Info:          *connectInfo,
		SentRequests:  make(map[string]int),
	}
	expiresAt := time.Now().Add(time.Duration(config.MainConfig.SessionTimeout) * time.Minute)
	tokenString, err := newToken(id, expiresAt)
	if err != nil {
		return "", time.Time{}, 0, err
	}

	s.sessionMap[id] = newSession
	s.sortedSessions = append(s.sortedSessions, newSession)
	s.count += 1 // increment session count
	queue, ok := types.NotifQueues[agent.CreatedBy]
	if ok {
		_ = queue.Enqueue(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelInfo,
			Msg:      fmt.Sprintf("new session from %s@%s", agent.Name, connectInfo.IPAddress),
		})
	}
	return tokenString, expiresAt, id, nil
}

func newToken(ID int, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, &Claims{
		ID: ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func validateJwt(c *http.Cookie) (*Claims, error) {
	tokenStr := c.Value
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return claims, err
	}
	if !token.Valid {
		return claims, fmt.Errorf("invalid token %s", tokenStr)
	}
	return claims, nil
}

// SessionByID retrieves an active HTTP connection with an agent, if said agent has ever had an active session
func (h *Handler) SessionByID(sessID int) *HTTPSession {
	return h.sessions.sessionMap[sessID]
}

// RmSession removes a session in special cases, such as if it dies unexpectedly
func (h *Handler) RmSession(sessID int) {
	h.sessions.lock.Lock()
	defer h.sessions.lock.Unlock()

	for i, ss := range h.sessions.sortedSessions {
		if ss.ID == sessID {
			h.sessions.sortedSessions = append(h.sessions.sortedSessions[:i], h.sessions.sortedSessions[i+1:]...)
		}
	}
	delete(h.sessions.sessionMap, sessID)
}
