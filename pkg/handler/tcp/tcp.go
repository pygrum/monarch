package tcp

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"flag"
	"github.com/pygrum/monarch/pkg/config"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/db"
	mhttp "github.com/pygrum/monarch/pkg/handler/http"
	"github.com/pygrum/monarch/pkg/log"
	"github.com/pygrum/monarch/pkg/protobuf/rpcpb"
	"github.com/pygrum/monarch/pkg/types"
	"io"
	"net"
	"strconv"
	"syscall"
	"time"
)

var (
	tl, fl        log.Logger
	MainHandler   *Handler
	ErrConnClosed = errors.New("tcp connection closed by client")
)

func Initialize() {
	tl, _ = log.NewLogger(log.TransientLogger, "")

	var err error
	fl, err = log.NewLogger(log.FileLogger, "tcp_handler")
	if err != nil {
		tl.Warn("couldn't initialize file logger: %v", err)
	}
	MainHandler, err = NewHandler()
	if err != nil {
		// this means we couldn't retrieve the tls config, so we die horribly
		tl.Fatal("%v", err)
	}
}

// Handler handles all raw TCP sessions with the same TLS configuration
type Handler struct {
	config   *tls.Config
	addr     string
	ln       net.Listener
	shutdown chan struct{}
	conn     chan net.Conn
	isActive bool
	sids     map[*net.Conn]*Conn
}

// Conn represents a single TCP session, which is translated and forwarded to an HTTP endpoint
type Conn struct {
	sid     int
	agent   *db.Agent
	session *mhttp.HTTPSession
}

func NewHandler() (*Handler, error) {
	addr := net.JoinHostPort(config.MainConfig.Interface, strconv.Itoa(config.MainConfig.TcpPort))
	tlsConfig, err := crypto.ServerTLSConfig()
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	h := &Handler{
		config:   tlsConfig,
		addr:     addr,
		shutdown: make(chan struct{}),
		conn:     make(chan net.Conn),
		sids:     make(map[*net.Conn]*Conn),
	}
	return h, nil
}

func (h *Handler) Stop() error {
	if h.ln == nil {
		return nil
	}
	err := h.ln.Close()
	h.ln = nil
	close(h.shutdown)
	return err
}

func (h *Handler) IsActive() bool {
	return h.ln != nil
}

func (h *Handler) Serve() {
	if h.ln != nil {
		return
	}
	ln, err := tls.Listen("tcp", h.addr, h.config)
	if err != nil {
		fl.Error("couldn't start listener: %v", err)
		return
	}
	h.ln = ln

	go h.Handle()
	go h.Accept()
}

func (h *Handler) Accept() {
	for {
		select {
		case <-h.shutdown:
			return
		default:
			conn, err := h.ln.Accept()
			if err != nil {
				fl.Error(err.Error())
				continue
			}
			h.conn <- conn
		}
	}
}

func (h *Handler) Handle() {
	for {
		select {
		case <-h.shutdown:
			return
		case conn := <-h.conn:
			go h.handleConn(conn)
		}
	}
}

func (h *Handler) handleConn(conn net.Conn) {
	buf, err := h.readPacket(conn)
	if err != nil {
		if errors.Is(err, ErrConnClosed) {
			fl.Info(err.Error())
			return
		}
		fl.Error("couldn't read from connection with %s: %v", conn.RemoteAddr().String(), err)
		conn.Close()
		return
	}
	// Register the new connection
	r, agent, err := ParseRegistration(buf)
	if err != nil {
		fl.Error("failed to parse registration: %v", err)
		conn.Close()
		return
	}
	r.IPAddress = conn.RemoteAddr().String()
	if flag.Lookup("test.v") != nil {
		return
	}
	_, _, id, err := mhttp.MainHandler.NewSession(agent, true, r)
	if err != nil {
		fl.Error("couldn't create new session: %v", err)
		conn.Close()
		return
	}
	ss := mhttp.MainHandler.SessionByID(id)
	if ss == nil {
		// something went horribly wrong during session creation
		// prevent a null deref
		fl.Error("nil session after creation")
		conn.Close()
		return
	}
	ss.LastActive = time.Now()
	ss.Status = mhttp.StatusActive
	c := &Conn{
		sid:     id,
		agent:   agent,
		session: ss,
	}
	h.sids[&conn] = c
	for {
		// blocking
		select {
		case <-ss.Killer:
			// send close if killed deliberately
			conn.Close()
			return
		case req := <-ss.RequestQueue.(*mhttp.RequestQueue).Channel:
			bytes, err := MarshalRequest(req)
			if err != nil {
				fl.Error("couldn't marshal request: %v", err)
				continue
			}
			// send request
			if _, err = conn.Write(bytes); err != nil {
				if IsConnClosedError(err) {
					delete(h.sids, &conn)
					mhttp.MainHandler.RmSession(c.sid)
					fl.Info(err.Error())
					conn.Close()
					return
				}
			}
			// inactive until we get a response
			ss.Status = mhttp.StatusInactive
			// blocking
			buf, err = h.readPacket(conn)
			if err != nil {
				if IsConnClosedError(err) {
					uid, err := db.GetIDByUsername(ss.UsedBy)
					if err == nil {
						queue, ok := types.NotifQueues[uid]
						if ok {
							queue.Enqueue(&rpcpb.Notification{
								LogLevel: rpcpb.LogLevel_LevelInfo,
								Msg:      ErrConnClosed.Error(),
							})
						}
					}
					// closure steps
					delete(h.sids, &conn)
					mhttp.MainHandler.RmSession(c.sid)
					_ = conn.Close()
				}
				fl.Error("couldn't read response from remote (%s): %v", conn.RemoteAddr().String(), err)
				continue
			}
			ss.Status = mhttp.StatusActive
			ss.LastActive = time.Now()
			r, err := ParseResponse(buf)
			if err != nil {
				fl.Error("couldn't parse response from remote: (%s): %v", conn.RemoteAddr().String(), err)
				continue
			}
			// queue for operator
			_ = ss.ResponseQueue.Enqueue(r)
		}
	}
}

func (h *Handler) readPacket(conn net.Conn) ([]byte, error) {
	s := make([]byte, 4)
	if _, err := conn.Read(s); err != nil {
		if IsConnClosedError(err) {
			c, ok := h.sids[&conn]
			if ok {
				mhttp.MainHandler.RmSession(c.sid)
				delete(h.sids, &conn)
				return nil, ErrConnClosed
			}
		}
		return nil, err
	}
	size := uint(binary.BigEndian.Uint32(s))
	buf := make([]byte, size)
	if _, err := conn.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func IsConnClosedError(err error) bool {
	switch {
	case
		errors.Is(err, net.ErrClosed),
		errors.Is(err, io.EOF),
		errors.Is(err, syscall.EPIPE):
		return true
	default:
		return false
	}
}
