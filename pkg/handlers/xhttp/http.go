package xhttp

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/transport"
	"net"
	"net/http"
	"strconv"
	"sync"
)

const (
	queueCapacity = 10
)

var (
	Handler *HTTPHandler
	l       log.Logger
)

type HTTPHandler struct {
	CertFile    string
	KeyFile     string
	httpServer  *http.Server
	httpsServer *http.Server
	httpsLock   bool
	httpLock    bool
	sessions    *sessions
}

func init() {
	Handler = NewHandler()
	l, _ = log.NewLogger(log.ConsoleLogger, "")
}

func (r *RQueue) Enqueue(req *transport.GenericHTTPRequest) error {
	select {
	case r.channel <- req:
		return nil
	default:
		return fmt.Errorf("queue is full - max capacity of %d\n", queueCapacity)
	}
}

func (r *RQueue) Dequeue() *transport.GenericHTTPRequest {
	select {
	case req := <-r.channel:
		return req
	default:
		return nil
	}
}

func (r *RQueue) Size() int {
	return len(r.channel)
}

func NewRQueue() *RQueue {
	return &RQueue{channel: make(chan *transport.GenericHTTPRequest, queueCapacity)}
}

func NewHandler() *HTTPHandler {
	ssns := &sessions{
		lock:       sync.Mutex{},
		sessionMap: make(map[int]*HTTPSession),
	}
	h := &HTTPHandler{
		CertFile: config.MainConfig.CertFile,
		KeyFile:  config.MainConfig.KeyFile,
		sessions: ssns,
	}
	router := mux.NewRouter()
	sRouter := mux.NewRouter()
	// Handles all requests
	router.PathPrefix("/").HandlerFunc(ssns.defaultHandler)
	sRouter.PathPrefix("/").HandlerFunc(ssns.defaultHandler)
	h.httpServer = &http.Server{
		Handler: router,
		Addr:    net.JoinHostPort(config.MainConfig.Interface, strconv.Itoa(config.MainConfig.HttpPort)),
	}
	h.httpsServer = &http.Server{
		Handler: sRouter,
		Addr:    net.JoinHostPort(config.MainConfig.Interface, strconv.Itoa(config.MainConfig.HttpsPort)),
	}
	return h
}

func (h *HTTPHandler) Serve() {
	// ensures can only be started once the server is available
	h.httpLock = true
	if err := h.httpServer.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		l.Error("listener failed: %v", err)
		return
	}
}

func (h *HTTPHandler) ServeTLS() {
	// ensures can only be started once the server is available
	h.httpsLock = true
	if err := h.httpsServer.ListenAndServeTLS(h.CertFile, h.KeyFile); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		l.Error("listener failed: %v", err)
	}
}

func (h *HTTPHandler) Stop() error {
	if err := h.httpServer.Shutdown(context.Background()); err != nil {
		return err
	}
	// create new server since shutdown destroys the old one
	h.httpServer = &http.Server{
		Handler: h.httpServer.Handler,
		Addr:    h.httpServer.Addr,
	}
	h.httpLock = false
	return nil
}

func (h *HTTPHandler) StopTLS() error {
	if err := h.httpsServer.Shutdown(context.Background()); err != nil {
		return err
	}
	h.httpsServer = &http.Server{
		Handler: h.httpsServer.Handler,
		Addr:    h.httpsServer.Addr,
	}
	h.httpsLock = false
	return nil
}

func (h *HTTPHandler) IsActive() bool {
	return h.httpLock
}

func (h *HTTPHandler) IsActiveTLS() bool {
	return h.httpsLock
}

func (h *HTTPHandler) QueueRequest(sessionID int, req *transport.GenericHTTPRequest) error {
	return h.sessions.sessionMap[sessionID].Queue.Enqueue(req) // returns error if queue is full
}

func (h *HTTPHandler) CancelRequest(sessionID int, req *transport.GenericHTTPRequest) error {
	if h.sessions.sessionMap[sessionID].Queue.Dequeue() == nil {
		return fmt.Errorf("no request to cancel")
	}
	return nil
}

func (h *HTTPHandler) Sessions(sessIDs []int) []*HTTPSession {
	var ss []*HTTPSession
	if len(sessIDs) == 0 {
		for _, v := range h.sessions.sessionMap {
			ss = append(ss, v)
		}
	} else {
		for _, sessID := range sessIDs {
			session := h.SessionByID(sessID)
			if session == nil {
				continue
			}
			ss = append(ss, session)
		}
	}
	return ss
}
