package transport

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"kama_chat_server/pkg/rpc/protocol"
)

var ErrConnClosed = errors.New("transport: connection is closed")

type OnFrameFunc func(conn *Conn, frame protocol.Frame)

type OnCloseFunc func(conn *Conn, err error)

type Conn struct {
	rawConn net.Conn

	writeChan chan protocol.Frame
	onFrame   OnFrameFunc
	onClose   OnCloseFunc

	closeOnce sync.Once
	sendMu    sync.RWMutex
	closed    bool
	closedCh  chan struct{}

	heartbeatMu       sync.RWMutex
	heartbeatEnabled  bool
	heartbeatInterval time.Duration
	heartbeatTimeout  time.Duration
	heartbeatStart    sync.Once

	healthMu     sync.RWMutex
	lastActiveAt time.Time
	lastPongAt   time.Time
}

func NewConn(rawConn net.Conn, onFrame OnFrameFunc, onClose OnCloseFunc) *Conn {
	now := time.Now()
	return &Conn{
		rawConn:      rawConn,
		writeChan:    make(chan protocol.Frame, 128),
		onFrame:      onFrame,
		onClose:      onClose,
		closedCh:     make(chan struct{}),
		lastActiveAt: now,
		lastPongAt:   now,
	}
}

func (conn *Conn) Start() {
	go conn.readLoop()
	go conn.writeLoop()
}

func (conn *Conn) WriteFrame(frame protocol.Frame) error {
	conn.sendMu.RLock()
	defer conn.sendMu.RUnlock()
	if conn.closed {
		return ErrConnClosed
	}
	select {
	case conn.writeChan <- frame:
		return nil
	case <-conn.closedCh:
		return ErrConnClosed
	}
}

func (conn *Conn) Close() error {
	return conn.shutdown(nil)
}

func (conn *Conn) RemoteAddr() net.Addr {
	return conn.rawConn.RemoteAddr()
}

func (conn *Conn) LastActiveAt() time.Time {
	conn.healthMu.RLock()
	defer conn.healthMu.RUnlock()
	return conn.lastActiveAt
}

func (conn *Conn) LastPongAt() time.Time {
	conn.healthMu.RLock()
	defer conn.healthMu.RUnlock()
	return conn.lastPongAt
}

func (conn *Conn) IsHealthy() bool {
	conn.sendMu.RLock()
	closed := conn.closed
	conn.sendMu.RUnlock()
	if closed {
		return false
	}

	conn.heartbeatMu.RLock()
	enabled := conn.heartbeatEnabled
	timeout := conn.heartbeatTimeout
	conn.heartbeatMu.RUnlock()
	if !enabled || timeout <= 0 {
		return true
	}

	return !conn.isHeartbeatExpired()
}

func (conn *Conn) readLoop() {
	for {
		frame, err := protocol.ReadFrame(conn.rawConn)
		if err != nil {
			_ = conn.shutdown(err)
			return
		}
		conn.markActive()

		switch frame.Header.MsgType {
		case protocol.MsgTypePing:
			_ = conn.WriteFrame(protocol.Frame{Header: protocol.FixedHeader{Version: protocol.CurrentVersion, MsgType: protocol.MsgTypePong, CodecID: protocol.DefaultCodecID, RequestID: frame.Header.RequestID}})
			continue
		case protocol.MsgTypePong:
			conn.markPong()
			continue
		}

		if conn.onFrame != nil {
			conn.onFrame(conn, frame)
		}
	}
}

func (conn *Conn) writeLoop() {
	for frame := range conn.writeChan {
		if err := protocol.WriteFrame(conn.rawConn, frame); err != nil {
			_ = conn.shutdown(err)
			return
		}
		conn.markActive()
	}
}

func (conn *Conn) shutdown(closeErr error) error {
	var connCloseErr error
	conn.closeOnce.Do(func() {
		conn.sendMu.Lock()
		conn.closed = true
		close(conn.writeChan)
		conn.sendMu.Unlock()

		close(conn.closedCh)
		connCloseErr = conn.rawConn.Close()
		if conn.onClose != nil {
			conn.onClose(conn, closeErr)
		}
	})
	if connCloseErr != nil {
		return fmt.Errorf("transport: close net connection: %w", connCloseErr)
	}
	return nil
}

func (conn *Conn) markActive() {
	conn.healthMu.Lock()
	defer conn.healthMu.Unlock()
	conn.lastActiveAt = time.Now()
}

func (conn *Conn) markPong() {
	conn.healthMu.Lock()
	defer conn.healthMu.Unlock()
	now := time.Now()
	conn.lastPongAt = now
	conn.lastActiveAt = now
}

func (conn *Conn) isHeartbeatExpired() bool {
	conn.heartbeatMu.RLock()
	timeout := conn.heartbeatTimeout
	conn.heartbeatMu.RUnlock()
	if timeout <= 0 {
		return false
	}
	conn.healthMu.RLock()
	lastPong := conn.lastPongAt
	conn.healthMu.RUnlock()
	return time.Since(lastPong) > timeout
}
