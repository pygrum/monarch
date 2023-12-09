package http

import (
	"fmt"
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

var (
	// new secret each restart
	key = []byte(uuid.New().String())
)

type HTTPSession struct {
	ID            int
	RequestQueue  Queue
	ResponseQueue Queue
	Agent         *db.Agent
	LastActive    time.Time
	Status        string
	lock          sync.Mutex
	Authenticated bool
	Info          transport.Registration
	SentRequests  map[string]int
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
func (s *sessions) newSession(agent *db.Agent, connectInfo *transport.Registration) (string, time.Time, int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	id := s.count
	status := ""
	// check if session with given agent exists anywhere
	for i, sess := range s.sortedSessions {
		if agent.AgentID == sess.Agent.AgentID {
			if !sess.Authenticated {
				return "", time.Time{}, 0, fmt.Errorf("agent %s was not previously authenticated", agent.AgentID)
			}
			delete(s.sessionMap, sess.ID) // remove session from map
			s.sortedSessions = append(s.sortedSessions[:i], s.sortedSessions[i+1:]...)
			status = "renewed"
		}
	}
	newSession := &HTTPSession{
		ID:            id,
		RequestQueue:  NewRequestQueue(),
		ResponseQueue: NewResponseQueue(),
		Agent:         agent,
		Info:          *connectInfo,
		Status:        status,
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
	queue, ok := NotifQueues[agent.CreatedBy]
	if ok {
		_ = queue.Enqueue(&rpcpb.Notification{
			LogLevel: rpcpb.LogLevel_LevelInfo,
			Msg:      fmt.Sprintf("new session from %s@%s (%s) \n", agent.Name, connectInfo.IPAddress, agent.AgentID),
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
