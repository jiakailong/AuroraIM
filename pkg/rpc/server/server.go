package server

import (
	"net"

	"kama_chat_server/pkg/rpc/protocol"
	"kama_chat_server/pkg/rpc/transport"
)

// Server 是 Day5 的 RPC Server MVP。
type Server struct {
	dispatcher *Dispatcher
	transport  *transport.Server
}

func NewServer() *Server {
	dispatcher := NewDispatcher()
	server := &Server{dispatcher: dispatcher}
	server.transport = transport.NewServer(func(conn *transport.Conn, frame protocol.Frame) {
		dispatcher.Dispatch(conn, frame)
	}, func(conn *transport.Conn, err error) {
		// Day5 暂无额外动作，连接清理由 transport 层处理。
	})
	return server
}

func (server *Server) Register(method string, handler HandlerFunc) {
	server.dispatcher.Register(method, handler)
}

func (server *Server) ListenAndServe(network, address string) error {
	return server.transport.ListenAndServe(network, address)
}

func (server *Server) Serve(listener net.Listener) error {
	return server.transport.Serve(listener)
}

func (server *Server) Close() error {
	return server.transport.Close()
}
