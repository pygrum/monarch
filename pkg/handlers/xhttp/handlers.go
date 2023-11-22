package xhttp

import (
	"encoding/json"
	"fmt"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/rpcpb"
	"github.com/pygrum/monarch/pkg/transport"
	"net/http"
	"os"
	"path/filepath"
)

func (s *sessions) defaultHandler(w http.ResponseWriter, r *http.Request) {
	first := &transport.GenericHTTPResponse{}
	if err := json.NewDecoder(r.Body).Decode(first); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	uuid := first.AgentID
	agent := &db.Agent{}
	if err := db.FindOneConditional("agent_id = ?", uuid, agent); err != nil || agent.AgentID == "" {
		// Just report as online
		w.WriteHeader(http.StatusOK)
		return
	}
	var sessionID int
	c, err := r.Cookie("PHPSESSID")
	if err != nil {
		// create new session and set cookie.
		// sessions can't be indexed by agent id otherwise there could be duplication
		token, expiresAt, id, err := s.newSession(agent)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sessionID = id
		c = &http.Cookie{
			Name:    "PHPSESSID",
			Expires: expiresAt,
			Value:   token,
			Secure:  true,
		}
		http.SetCookie(w, c)
	} else {
		if c == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// verify JWT
		claims, err := validateJwt(c)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		sessionID = claims.ID
	}
	var resp *transport.GenericHTTPRequest
	// should always return a session
	session, ok := s.sessionMap[sessionID]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	handleResponse(session, first)
	for {
		// Keep checking for new response
		resp = session.Queue.Dequeue()
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

func handleResponse(session *HTTPSession, resp *transport.GenericHTTPResponse) {
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
