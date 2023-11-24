package xhttp

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/pygrum/monarch/pkg/coop"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/transport"
	"net/http"
	"sync"
	"time"
)

var (
	// new secret each restart
	key = []byte(uuid.New().String())
)

type HTTPSession struct {
	ID         int
	Queue      *RQueue
	Agent      *db.Agent
	LastActive time.Time
	Player     *coop.Player // nil if console is using it
	lock       sync.Mutex
}

// RQueue holds up to queueCapacity responses for a callback.
// If full, an error is raised.
type RQueue struct {
	channel chan *transport.GenericHTTPRequest
}

type Claims struct {
	jwt.RegisteredClaims
	ID int
}

type sessions struct {
	lock       sync.Mutex
	count      int
	sessionMap map[int]*HTTPSession
}

// newSession registers a session and creates a cookie for auth
func (s *sessions) newSession(agent *db.Agent) (string, time.Time, int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	id := s.count
	newSession := &HTTPSession{
		Queue: NewRQueue(),
		Agent: agent,
	}
	expiresAt := time.Now().Add(5 * time.Minute)
	tokenString, err := newToken(id, expiresAt)
	if err != nil {
		return "", time.Time{}, 0, err
	}
	s.sessionMap[id] = newSession
	s.count += 1 // increment session count
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
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token %s", tokenStr)
	}
	return claims, nil
}

// SessionByID retrieves an active HTTP connection with an agent, if said agent has ever had an active session
func (h *HTTPHandler) SessionByID(sessID int) *HTTPSession {
	return h.sessions.sessionMap[sessID]
}
