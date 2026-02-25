package transport

import (
	"errors"
	"time"

	"kama_chat_server/pkg/rpc/protocol"
)

var ErrHeartbeatTimeout = errors.New("transport: heartbeat timeout")

func (conn *Conn) EnableHeartbeat(interval, timeout time.Duration) error {
	if interval <= 0 || timeout <= 0 || timeout < interval {
		return errors.New("transport: invalid heartbeat config")
	}

	conn.heartbeatMu.Lock()
	conn.heartbeatEnabled = true
	conn.heartbeatInterval = interval
	conn.heartbeatTimeout = timeout
	conn.heartbeatMu.Unlock()
	conn.markPong()

	conn.heartbeatStart.Do(func() {
		go conn.heartbeatLoop()
	})
	return nil
}

func (conn *Conn) heartbeatLoop() {
	conn.heartbeatMu.RLock()
	interval := conn.heartbeatInterval
	conn.heartbeatMu.RUnlock()
	if interval <= 0 {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-conn.closedCh:
			return
		case <-ticker.C:
			if conn.isHeartbeatExpired() {
				_ = conn.shutdown(ErrHeartbeatTimeout)
				return
			}

			pingFrame := protocol.Frame{Header: protocol.FixedHeader{Version: protocol.CurrentVersion, MsgType: protocol.MsgTypePing, CodecID: protocol.DefaultCodecID, RequestID: uint64(time.Now().UnixNano())}}
			if err := conn.WriteFrame(pingFrame); err != nil {
				_ = conn.shutdown(err)
				return
			}
		}
	}
}
