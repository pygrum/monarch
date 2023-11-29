package xhttp

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/rpcpb"
	"github.com/pygrum/monarch/pkg/transport"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	cookieName = "token"
)

// http://host:port/login
func (s *sessions) loginHandler(w http.ResponseWriter, r *http.Request) {
	connectInfo := &transport.Registration{}
	defer r.Body.Close()
	connectInfoBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(connectInfoBytes, connectInfo); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uuid := connectInfo.AgentID
	agent := &db.Agent{}
	if err := db.FindOneConditional("agent_id = ?", uuid, agent); err != nil || agent.AgentID == "" {
		fl.Error("couldn't find agent '%s': %v", uuid, err)
		// Just report as online
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token, expiresAt, _, err := s.newSession(agent, connectInfo)
	if err != nil {
		fl.Error("failed to create new session: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	c, err := r.Cookie(cookieName)
	c = &http.Cookie{
		Name:    cookieName,
		Expires: expiresAt,
		Value:   token,
		Secure:  true,
	}
	// allows agent to use cookie from first request for subsequent ones
	http.SetCookie(w, c)
	w.WriteHeader(http.StatusOK)
	l.Info("new session from %s - id: %s\n", r.RemoteAddr, agent.AgentID)
}

// http://host:port/
func (s *sessions) defaultHandler(w http.ResponseWriter, r *http.Request) {
	var sessionID int
	c, err := r.Cookie(cookieName)
	if err != nil || c == nil {
		fl.Error("%v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	response := &transport.GenericHTTPResponse{}
	claims, err := validateJwt(c) // is invalid after server restart
	if err != nil {
		fl.Error("jwt validation failed: %v", err)
		// if there was a leftover response from an expired session, queue it anyway
		// kinda dangerous if there was no initial request, so we should verify there's a request with a matching ID
		if errors.Is(err, jwt.ErrTokenExpired) {
			if err = json.NewDecoder(r.Body).Decode(response); err != nil {
				fl.Error("failed to parse response from %s: %v", r.RemoteAddr, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			ss := s.sessionMap[claims.ID]
			// ingest response despite expiry
			if _, ok := ss.SentRequests[response.RequestID]; ok {
				_ = ss.ResponseQueue.Enqueue(response)
				delete(ss.SentRequests, response.RequestID)
			}
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	sessionID = claims.ID
	var resp *transport.GenericHTTPRequest
	// should always return a session
	session, ok := s.sessionMap[sessionID]
	// potentially won't let someone re-auth if server goes down despite having valid cookie
	if !ok {
		fl.Error("session '%d' not found", sessionID)
		// use status bad request to tell client to ditch the cookie, since the server restarted
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Refresh token if there's less than 2.5 minutes until expiry
	if claims.ExpiresAt.Sub(time.Now()) < (150 * time.Second) {
		expiresAt := time.Now().Add(10 * time.Minute)
		tok, err := newToken(claims.ID, expiresAt)
		if err == nil {
			c = &http.Cookie{
				Name:    cookieName,
				Expires: expiresAt,
				Value:   tok,
				Secure:  true,
			}
			http.SetCookie(w, c)
		}
	}
	// session is authenticated if JWT has been validated
	if !session.Authenticated {
		session.Authenticated = true
		session.Status = "active"
		session.LastActive = time.Now()
	} else {
		if r.Body == http.NoBody {
			fl.Error("received empty body during authenticated session")
			w.WriteHeader(http.StatusOK)
			return
		}
		if err = json.NewDecoder(r.Body).Decode(response); err != nil {
			fl.Error("failed to parse response from %s: %v", r.RemoteAddr, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Queue the message as a response since this is not the first authenticated message
		_ = s.sessionMap[sessionID].ResponseQueue.Enqueue(response)
		HandleResponse(session, response)
	}
	for {
		// Keep checking for new request (blocking)
		resp = session.RequestQueue.Dequeue().(*transport.GenericHTTPRequest)
		if resp != nil {
			// Then someone queued request, so send it
			b, err := json.Marshal(resp)
			if err != nil {
				fl.Error("marshalling request %s failed: %v", ShortID(resp.RequestID), err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
			// add to sent requests, the number doesn't matter
			session.SentRequests[resp.RequestID] = 0
			return // exit so we can get a response
		}
	}
}

func HandleResponse(session *HTTPSession, resp *transport.GenericHTTPResponse) {
	session.LastActive = time.Now()
	for _, response := range resp.Responses {
		handleResponse(session, response, ShortID(resp.RequestID))
	}
}

func handleResponse(session *HTTPSession, response transport.ResponseDetail, rid string) {
	if response.Status == rpcpb.Status_FailedWithMessage {
		// Will definitely be a console player as we don't have multiplayer yet so no need to add the clause
		if session.Player.ConsolePlayer() {
			if len(response.Data) == 0 {
				l.Error("request %s failed but no message was returned", rid)
				return
			}
			l.Error("%s failed: %s", rid, string(response.Data))
			return
		}
	}
	if response.Dest == transport.DestStdout {
		if session.Player.ConsolePlayer() {
			log.Print(string(response.Data))
		}
	} else if response.Dest == transport.DestFile {
		file := filepath.Join(os.TempDir(), response.Name)
		if err := os.WriteFile(file, response.Data, 0666); err != nil {
			if session.Player.ConsolePlayer() {
				l.Error("failed writing response to %s to file: %v", rid, err)
			}
			return
		}
		if session.Player.ConsolePlayer() {
			l.Info("%s: file saved to %s", rid, file)
		}
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fl.Info("remote=%s method=%s url=%s content-length=%d", r.RemoteAddr,
			r.Method, r.URL.String(), r.ContentLength)
		next.ServeHTTP(w, r)
	})
}

func ShortID(uuid string) string {
	return uuid[:8]
}
