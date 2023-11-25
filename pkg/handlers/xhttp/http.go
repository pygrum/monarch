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
	fl      log.Logger
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

type Queue interface {
	Enqueue(interface{}) error
	Dequeue() interface{}
	Size() int
}

// RequestQueue holds up to queueCapacity responses for a callback.
// If full, an error is raised.
type RequestQueue struct {
	channel chan *transport.GenericHTTPRequest
}

type ResponseQueue struct {
	channel chan *transport.GenericHTTPResponse
}

func init() {
	Handler = NewHandler()
	l, _ = log.NewLogger(log.ConsoleLogger, "")

	var err error
	fl, err = log.NewLogger(log.FileLogger, "handler")
	if err != nil {
		l.Warn("could not create file logger: %v", err)
	}
}

func (r *RequestQueue) Enqueue(req interface{}) error {
	select {
	case r.channel <- req.(*transport.GenericHTTPRequest):
		return nil
	default:
		return fmt.Errorf("queue is full - max capacity of %d\n", queueCapacity)
	}
}

func (r *RequestQueue) Dequeue() interface{} {
	// Must block, as we wait for a request to queue
	select {
	case req := <-r.channel:
		return req
	}
}

func (r *RequestQueue) Size() int {
	return len(r.channel)
}

func (r *ResponseQueue) Enqueue(req interface{}) error {
	select {
	case r.channel <- req.(*transport.GenericHTTPResponse):
		return nil
	default:
		return fmt.Errorf("queue is full - max capacity of %d\n", queueCapacity)
	}
}

func (r *ResponseQueue) Dequeue() interface{} {
	// Must block, as we wait for a request to queue
	select {
	case req := <-r.channel:
		return req
	}
}

func (r *ResponseQueue) Size() int {
	return len(r.channel)
}

func NewRequestQueue() *RequestQueue {
	return &RequestQueue{channel: make(chan *transport.GenericHTTPRequest, queueCapacity)}
}

func NewResponseQueue() *ResponseQueue {
	return &ResponseQueue{channel: make(chan *transport.GenericHTTPResponse, queueCapacity)}
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
	ss := h.sessions.sessionMap[sessionID]
	if ss == nil {
		return fmt.Errorf("session '%d' no longer exists - it may have expired due to a new connection",
			sessionID)
	}
	return ss.RequestQueue.Enqueue(req) // returns error if queue is full
}

func (h *HTTPHandler) AwaitResponse(sessionID int) *transport.GenericHTTPResponse {
	// returns error if queue is full
	return h.sessions.sessionMap[sessionID].ResponseQueue.Dequeue().(*transport.GenericHTTPResponse)
}

func (h *HTTPHandler) Sessions(sessIDs []int) []*HTTPSession {
	h.sessions.lock.Lock()
	defer h.sessions.lock.Unlock()
	var ss []*HTTPSession
	if len(sessIDs) == 0 {
		return h.sessions.sortedSessions
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
