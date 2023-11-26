package xhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/rpcpb"
	"github.com/pygrum/monarch/pkg/transport"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (s *sessions) defaultHandler(w http.ResponseWriter, r *http.Request) {
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
	var sessionID int
	c, err := r.Cookie("PHPSESSID")
	if err != nil || c == nil {
		// create new session and set cookie.
		// sessions can't be indexed by agent id otherwise there could be duplication
		token, expiresAt, id, err := s.newSession(agent, connectInfo)
		if err != nil {
			fl.Error("failed to create new session: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		sessionID = id
		c = &http.Cookie{
			Name:    "PHPSESSID",
			Expires: expiresAt,
			Value:   token,
			Secure:  true,
		}
		// allows agent to use cookie from first request for subsequent ones
		http.SetCookie(w, c)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		claims, err := validateJwt(c) // is invalid after server restart
		if err != nil {
			fl.Error("jwt validation failed: %v", err)
			// if there was a leftover response from an expired session, queue it anyway
			// kinda dangerous if there was no initial request, so we should verify there's a request with a matching ID
			if errors.Is(err, jwt.ErrTokenExpired) {
				if connectInfo.Data != nil {
					rid := connectInfo.Data.RequestID
					sent := s.sessionMap[sessionID].SentRequests
					// check if there's a corresponding request
					_, ok := sent[rid]
					if !ok {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					// ingest response even if unauthenticated
					_ = s.sessionMap[sessionID].ResponseQueue.Enqueue(connectInfo.Data)
					delete(s.sortedSessions[sessionID].SentRequests, rid)
				}
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		sessionID = claims.ID
	}
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
	// session is authenticated if JWT has been validated
	if !session.Authenticated {
		session.Authenticated = true
		session.Status = "active"
		session.LastActive = time.Now()
	} else {
		// must be a response to an issued request since the agent is already authenticated
		response := &transport.GenericHTTPResponse{}
		if err = json.Unmarshal(connectInfoBytes, response); err != nil {
			fl.Error("failed to parse response from agent %s: %v", connectInfo.AgentID, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Queue the message as a response since this is not the first authenticated message
		_ = session.ResponseQueue.Enqueue(response)
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
		handleResponseDetails(session, response, ShortID(resp.RequestID))
	}
	if session.Player.ConsolePlayer() {
		fmt.Println()
	}
}

func handleResponseDetails(session *HTTPSession, response transport.ResponseDetail, rid string) {
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
			fmt.Println(string(response.Data))
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
		fl.Info("method=%s url=%s content-length=%d", r.Method, r.URL.String(), r.ContentLength)
		next.ServeHTTP(w, r)
	})
}

func ShortID(uuid string) string {
	return uuid[:8]
}
