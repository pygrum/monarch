package xhttp

import (
	"encoding/json"
	"fmt"
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
	first := &transport.GenericHTTPResponse{}
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(b, first); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uuid := first.AgentID
	agent := &db.Agent{}
	if err := db.FindOneConditional("agent_id = ?", uuid, agent); err != nil || agent.AgentID == "" {
		// Just report as online
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var sessionID int
	c, err := r.Cookie("PHPSESSID")
	if err != nil || c == nil {
		// create new session and set cookie.
		// sessions can't be indexed by agent id otherwise there could be duplication
		token, expiresAt, id, err := s.newSession(agent)
		if err != nil {
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
		// use status bad request to tell client to ditch the cookie, since the server restarted
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// session is authenticated if JWT has been validated
	if !session.Authenticated {
		session.Authenticated = true
	} else {
		// Queue the message as a response since this is not the first authenticated message
		_ = session.ResponseQueue.Enqueue(first)
	}
	for {
		// Keep checking for new response
		resp = session.RequestQueue.Dequeue().(*transport.GenericHTTPRequest)
		if resp != nil {
			// Then someone queued request, so send it
			b, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(b)
			return // exit so we can get a response
		}
	}
	// TODO: DO something with the request body. save it to an agent-specific history file or something
}

func HandleResponse(session *HTTPSession, resp *transport.GenericHTTPResponse) {
	session.LastActive = time.Now()
	if resp.Status == rpcpb.Status_FailedWithMessage {
		// Will definitely be a console player as we don't have multiplayer yet so no need to add the clause
		if session.Player.ConsolePlayer() {
			if len(resp.Responses) == 0 {
				l.Error("request %s failed but no message was returned", resp.RequestID)
				return
			}
			l.Error("%s failed: %s", ShortID(resp.RequestID), string(resp.Responses[0].Data))
			return
		}
	} else {
		for _, response := range resp.Responses {
			if response.Dest == transport.DestStdout {
				if session.Player.ConsolePlayer() {
					fmt.Println(string(response.Data))
				}
			} else if response.Dest == transport.DestFile {
				file := filepath.Join(os.TempDir(), response.Name)
				if err := os.WriteFile(file, response.Data, 0666); err != nil {
					if session.Player.ConsolePlayer() {
						l.Error("failed writing response to %s to file: %v", ShortID(resp.RequestID), err)
					}
					return
				}
				if session.Player.ConsolePlayer() {
					l.Info("%s: file saved to %s", ShortID(resp.RequestID), file)
				}
			}
		}
	}
}

func ShortID(uuid string) string {
	return uuid[:8]
}
