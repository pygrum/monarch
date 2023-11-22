package xhttp

import (
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
	CertFile string
	KeyFile  string
	sessions *sessions
	Router   *mux.Router
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
	r := mux.NewRouter()
	// Handles all requests
	r.PathPrefix("/").HandlerFunc(ssns.defaultHandler)
	h.Router = r
	return h
}

func (h *HTTPHandler) Serve(iface string) error {
	if len(iface) == 0 {
		iface = config.MainConfig.Interface
	}
	server := http.Server{
		Handler: h.Router,
		Addr:    net.JoinHostPort(iface, strconv.Itoa(config.MainConfig.HttpsPort)),
	}
	return server.ListenAndServe()
}

func (h *HTTPHandler) ServeTLS(iface string) error {
	if len(iface) == 0 {
		iface = config.MainConfig.Interface
	}
	server := http.Server{
		Handler: h.Router,
		Addr:    net.JoinHostPort(iface, strconv.Itoa(config.MainConfig.HttpsPort)),
	}
	return server.ListenAndServeTLS(h.CertFile, h.KeyFile)
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
	for _, sessID := range sessIDs {
		session := h.SessionByID(sessID)
		if session == nil {
			continue
		}
		ss = append(ss, session)
	}
	return ss
}
