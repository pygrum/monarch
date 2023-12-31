package http

import (
	"errors"
	"fmt"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/db"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/transport"
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
	CertFile    string
	KeyFile     string
	httpServer  *http.Server
	httpsServer *http.Server
	isActiveTLS bool
	isActive    bool
	sessions    *sessions
}

// RequestQueue holds up to queueCapacity responses for a callback.
// If full, an error is raised.
type RequestQueue struct {
	// expose for TCP
	Channel chan *transport.GenericHTTPRequest
}

type ResponseQueue struct {
	channel chan *transport.GenericHTTPResponse
}

func Initialize() {
	MainHandler = NewHandler()
	TranLogger, _ = log.NewLogger(log.TransientLogger, "")

	var err error
	fl, err = log.NewLogger(log.FileLogger, "handler")
	if err != nil {
		TranLogger.Warn("could not create file logger: %v", err)
	}
}

func ClientInitialize() {
	TranLogger, _ = log.NewLogger(log.TransientLogger, "")
}

func (r *RequestQueue) Enqueue(req interface{}) error {
	select {
	case r.Channel <- req.(*transport.GenericHTTPRequest):
		return nil
	default:
		return fmt.Errorf("queue is full - max capacity of %d\n", queueCapacity)
	}
}

func (r *RequestQueue) Dequeue() interface{} {
	// Must block, as we wait for a request to queue
	select {
	case req := <-r.Channel:
		return req
	}
}

func (r *RequestQueue) Size() int {
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
	return &RequestQueue{Channel: make(chan *transport.GenericHTTPRequest, queueCapacity)}
}

func NewResponseQueue() *ResponseQueue {
	return &ResponseQueue{channel: make(chan *transport.GenericHTTPResponse, queueCapacity)}
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
	h.isActive = false
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
	h.isActiveTLS = false
	return nil
}

func (h *Handler) IsActive() bool {
	return h.isActive
}

func (h *Handler) IsActiveTLS() bool {
	return h.isActiveTLS
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

func (h *Handler) NewSession(agent *db.Agent, isTCP bool, connectInfo *transport.Registration) (string, time.Time, int, error) {
	return h.sessions.newSession(agent, isTCP, connectInfo)
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
		CertFile: config.MainConfig.CertFile,
		KeyFile:  config.MainConfig.KeyFile,
		sessions: ssns,
	}
	router := mux.NewRouter()
	sRouter := mux.NewRouter()

	h.setupRouter(router)
	h.setupRouter(sRouter)

	h.httpServer = &http.Server{
		Handler: router,
		Addr:    net.JoinHostPort(config.MainConfig.Interface, strconv.Itoa(config.MainConfig.HttpPort)),
	}
	tlsConfig, err := crypto.ServerTLSConfig()
	if err != nil {
		fl.Fatal("couldn't get server TLS config: %v", err)
	}
	h.httpsServer = &http.Server{
		Handler:   sRouter,
		TLSConfig: tlsConfig,
		Addr:      net.JoinHostPort(config.MainConfig.Interface, strconv.Itoa(config.MainConfig.HttpsPort)),
	}
	return h
}

func (h *Handler) setupRouter(router *mux.Router) {
	for _, path := range config.C2Config.LoginEndpoint.Paths {
		router.HandleFunc(path.Path, h.sessions.loginHandler).Methods(path.Methods...)
	}
	for _, path := range config.C2Config.StageEndpoint.Paths {
		router.HandleFunc(path.Path, h.sessions.stageHandler).Methods(path.Methods...)
	}
	for _, path := range config.C2Config.MainEndpoint.Paths {
		router.PathPrefix(path.Path).HandlerFunc(h.sessions.defaultHandler).Methods(path.Methods...)
	}
	router.Use(loggingMiddleware)
}

func (h *Handler) Serve() {
	lh, err := net.Listen("tcp", h.httpServer.Addr)
	if err != nil {
		fl.Error("failed to initialise listening socket: %v", err)
		return
	}
	// ensures can only be started once the TeamServer is available
	h.isActive = true
	if err = h.httpServer.Serve(lh); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		fl.Error("listener failed: %v", err)
		return
	}
}

func (h *Handler) ServeTLS() {
	lh, err := net.Listen("tcp", h.httpsServer.Addr)
	if err != nil {
		fl.Error("failed to initialise listening socket: %v", err)
		return
	}
	// ensures can only be started once the server is available
	h.isActiveTLS = true
	if err = h.httpsServer.ServeTLS(lh, h.CertFile, h.KeyFile); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		fl.Error("listener failed: %v", err)
	}
}
