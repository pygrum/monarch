package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/protobuf/builderpb"
	"github.com/pygrum/monarch/pkg/protobuf/clientpb"
	"github.com/pygrum/monarch/pkg/transport"
)

const (
	cookieName = "token"
)

// http://host:port/index/{file}
func stageHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	params := mux.Vars(r)
	file, ok := params["file"]
	if ok {
		data, err := Stage.get(file)
		if err != nil {
			fl.Error(err.Error())
			return
		}
		w.Write(data)
	}
}

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
	connectInfo.IPAddress = r.RemoteAddr
	token, expiresAt, _, err := s.newSession(agent, false, connectInfo)
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
	claims, err := validateJwt(c) // is invalid after team-server restart
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
	// potentially won't let someone re-auth if teamserver goes down despite having valid cookie
	if !ok {
		fl.Error("session '%d' not found", sessionID)
		// use status bad request to tell client to ditch the cookie, since the teamserver restarted
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	session.Status = StatusActive
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
		session.Status = StatusActive
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
	}
	for {
		select {
		case <-r.Context().Done():
			session.Status = StatusKilled
			agent := session.Agent
			queue, ok := types.NotifQueues[agent.CreatedBy]
			if ok {
				_ = queue.Enqueue(&rpcpb.Notification{
					LogLevel: rpcpb.LogLevel_LevelWarn,
					Msg:      fmt.Sprintf("session %d (%s) died unexpectedly", session.ID, session.Agent.Name),
				})
			}
			return
		case resp = <-session.RequestQueue.(*RequestQueue).channel:
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
				session.Status = StatusInactive // will be active again after callback
				return                          // exit so we can get a response
			}
		}
	}
}

func HandleResponse(session *clientpb.Session, resp *transport.GenericHTTPResponse) {
	session.LastActive = time.Now().Format(time.RFC850)
	for _, response := range resp.Responses {
		handleResponse(response, ShortID(resp.RequestID))
	}
}

func handleResponse(response transport.ResponseDetail, rid string) {
	if response.Status == builderpb.Status_FailedWithMessage {
		if len(response.Data) == 0 {
			TranLogger.Error("request %s failed but no message was returned", rid)
			return
		}
		TranLogger.Error("%s failed: %s", rid, string(response.Data))
		return
	}
	if response.Dest == transport.DestStdout {
		log.Print(string(response.Data))
	} else if response.Dest == transport.DestFile {
		wd, err := os.Getwd()
		if err != nil {
			wd = os.TempDir()
		}
		file := filepath.Join(wd, response.Name)
		if err := os.WriteFile(file, response.Data, 0666); err != nil {
			TranLogger.Error("failed writing response to %s to file: %v", rid, err)
			return
		}
		TranLogger.Info("%s: file saved to %s", rid, file)
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
