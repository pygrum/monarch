package http

import (
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
	MainHandler *Handler
	TranLogger  log.Logger
	fl          log.Logger
)

type Handler struct {
	CertFile             string
	KeyFile              string
	httpServer           *http.Server
	httpsServer          *http.Server
	httpsLock            bool
	httpLock             bool
	SessionNotifications Queue
	sessions             *sessions
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

type SessInfoQueue struct {
	Channel chan transport.Registration
}

func init() {
	MainHandler = NewHandler()
	TranLogger, _ = log.NewLogger(log.TransientLogger, "")

	var err error
	fl, err = log.NewLogger(log.FileLogger, "handler")
	if err != nil {
		TranLogger.Warn("could not create file logger: %v", err)
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

func (r *SessInfoQueue) Enqueue(req interface{}) error {
	select {
	case r.Channel <- req.(transport.Registration):
		return nil
	default:
		return fmt.Errorf("queue is full - max capacity of %d\n", queueCapacity)
	}
}

func (r *SessInfoQueue) Dequeue() interface{} {
	// Must block, as we wait for a request to queue
	select {
	case req := <-r.Channel:
		return req
	}
}

func (r *SessInfoQueue) Size() int {
	return len(r.Channel)
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

func NewSessInfoQueue() *SessInfoQueue {
	return &SessInfoQueue{Channel: make(chan transport.Registration, queueCapacity)}
}

func (h *Handler) Stop() error {
	if err := h.httpServer.Close(); err != nil {
		return err
	}
	// create new team-server since old one is destroyed
	h.httpServer = &http.Server{
		Handler: h.httpServer.Handler,
		Addr:    h.httpServer.Addr,
	}
	h.httpLock = false
	return nil
}

func (h *Handler) StopTLS() error {
	if err := h.httpServer.Close(); err != nil {
		return err
	}
	h.httpsServer = &http.Server{
		Handler: h.httpsServer.Handler,
		Addr:    h.httpsServer.Addr,
	}
	h.httpsLock = false
	return nil
}

func (h *Handler) IsActive() bool {
	return h.httpLock
}

func (h *Handler) IsActiveTLS() bool {
	return h.httpsLock
}

func (h *Handler) QueueRequest(sessionID int, req *transport.GenericHTTPRequest) error {
	ss := h.sessions.sessionMap[sessionID]
	if ss == nil {
		return fmt.Errorf("session '%d' no longer exists - it may have expired due to a new connection",
			sessionID)
	}
	return ss.RequestQueue.Enqueue(req) // returns error if queue is full
}

func (h *Handler) AwaitResponse(sessionID int) *transport.GenericHTTPResponse {
	// returns error if queue is full
	return h.sessions.sessionMap[sessionID].ResponseQueue.Dequeue().(*transport.GenericHTTPResponse)
}

func (h *Handler) Sessions(sessIDs []int) []*HTTPSession {
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

func NewHandler() *Handler {
	ssns := &sessions{
		lock:       sync.Mutex{},
		sessionMap: make(map[int]*HTTPSession),
	}
	h := &Handler{
		CertFile:             config.MainConfig.CertFile,
		KeyFile:              config.MainConfig.KeyFile,
		SessionNotifications: NewSessInfoQueue(),
		sessions:             ssns,
	}
	router := mux.NewRouter()
	sRouter := mux.NewRouter()
	mEp := config.MainConfig.MainEndpoint
	lEp := config.MainConfig.LoginEndpoint
	sEp := config.MainConfig.StageEndpoint
	router.HandleFunc(lEp, ssns.loginHandler)
	router.HandleFunc(sEp, stageHandler)
	// Handles all requests
	router.PathPrefix(mEp).HandlerFunc(ssns.defaultHandler)
	router.Use(loggingMiddleware)

	sRouter.HandleFunc(lEp, ssns.loginHandler)
	sRouter.HandleFunc(sEp, stageHandler)
	sRouter.PathPrefix(mEp).HandlerFunc(ssns.defaultHandler)
	sRouter.Use(loggingMiddleware)

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

func (h *Handler) Serve() {
	lh, err := net.Listen("tcp", h.httpServer.Addr)
	if err != nil {
		TranLogger.Error("failed to initialise listening socket: %v", err)
		return
	}
	// ensures can only be started once the teamserver is available
	h.httpLock = true
	if err = h.httpServer.Serve(lh); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		TranLogger.Error("listener failed: %v", err)
		return
	}
}

func (h *Handler) ServeTLS() {
	lh, err := net.Listen("tcp", h.httpsServer.Addr)
	if err != nil {
		TranLogger.Error("failed to initialise listening socket: %v", err)
		return
	}
	// ensures can only be started once the teamserver is available
	h.httpsLock = true
	if err = h.httpsServer.ServeTLS(lh, h.CertFile, h.KeyFile); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		TranLogger.Error("listener failed: %v", err)
	}
}
