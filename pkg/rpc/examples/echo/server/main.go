package main

import (
	"flag"
	"fmt"
	"log"

	"kama_chat_server/pkg/rpc/protocol"
	"kama_chat_server/pkg/rpc/transport"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:19090", "echo server listen address")
	flag.Parse()

	server := transport.NewServer(func(conn *transport.Conn, frame protocol.Frame) {
		if err := conn.WriteFrame(frame); err != nil {
			log.Printf("echo write failed: remote=%s err=%v", conn.RemoteAddr(), err)
		}
	}, func(conn *transport.Conn, err error) {
		if err != nil {
			log.Printf("connection closed: remote=%s err=%v", conn.RemoteAddr(), err)
			return
		}
		log.Printf("connection closed: remote=%s", conn.RemoteAddr())
	})

	fmt.Printf("echo server listening on %s\n", *addr)
	if err := server.ListenAndServe("tcp", *addr); err != nil && err != transport.ErrServerClosed {
		log.Fatalf("echo server stopped with error: %v", err)
	}
}
