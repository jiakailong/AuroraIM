package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	rpcerrors "kama_chat_server/pkg/rpc/errors"
	rpcserver "kama_chat_server/pkg/rpc/server"
)

type MessageService struct {
	mu sync.Mutex

	nextMessageID uint64
	offlineByUser map[string][]Message
	dedupIndex    map[string]SendMessageResponse
}

type Message struct {
	MessageID   uint64 `json:"message_id"`
	FromUserID  string `json:"from_user_id"`
	ToUserID    string `json:"to_user_id"`
	Content     string `json:"content"`
	ClientMsgID string `json:"client_msg_id"`
	SentAtUnix  int64  `json:"sent_at_unix"`
}

type SendMessageRequest struct {
	FromUserID  string `json:"from_user_id"`
	ToUserID    string `json:"to_user_id"`
	Content     string `json:"content"`
	ClientMsgID string `json:"client_msg_id"`
}

type SendMessageResponse struct {
	AckMessageID uint64 `json:"ack_message_id"`
	Duplicated   bool   `json:"duplicated"`
}

type PullOfflineRequest struct {
	UserID string `json:"user_id"`
	Limit  int    `json:"limit"`
}

type PullOfflineResponse struct {
	Messages []Message `json:"messages"`
}

func NewMessageService() *MessageService {
	return &MessageService{
		nextMessageID: 1,
		offlineByUser: make(map[string][]Message),
		dedupIndex:    make(map[string]SendMessageResponse),
	}
}

func (service *MessageService) SendMessage(ctx context.Context, request *SendMessageRequest) (*SendMessageResponse, error) {
	if request == nil {
		return nil, rpcerrors.New(rpcerrors.BadRequest, "request is nil", nil)
	}
	if request.FromUserID == "" || request.ToUserID == "" {
		return nil, rpcerrors.New(rpcerrors.BadRequest, "from_user_id and to_user_id are required", nil)
	}
	if request.Content == "" {
		return nil, rpcerrors.New(rpcerrors.BadRequest, "content is required", nil)
	}
	if request.ClientMsgID == "" {
		return nil, rpcerrors.New(rpcerrors.BadRequest, "client_msg_id is required", nil)
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	dedupKey := request.FromUserID + ":" + request.ClientMsgID
	if existing, ok := service.dedupIndex[dedupKey]; ok {
		response := existing
		response.Duplicated = true
		return &response, nil
	}

	message := Message{
		MessageID:   service.nextMessageID,
		FromUserID:  request.FromUserID,
		ToUserID:    request.ToUserID,
		Content:     request.Content,
		ClientMsgID: request.ClientMsgID,
		SentAtUnix:  time.Now().Unix(),
	}
	service.nextMessageID++
	service.offlineByUser[request.ToUserID] = append(service.offlineByUser[request.ToUserID], message)

	response := SendMessageResponse{AckMessageID: message.MessageID, Duplicated: false}
	service.dedupIndex[dedupKey] = response
	return &response, nil
}

func (service *MessageService) PullOffline(ctx context.Context, request *PullOfflineRequest) (*PullOfflineResponse, error) {
	if request == nil {
		return nil, rpcerrors.New(rpcerrors.BadRequest, "request is nil", nil)
	}
	if request.UserID == "" {
		return nil, rpcerrors.New(rpcerrors.BadRequest, "user_id is required", nil)
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	queued := service.offlineByUser[request.UserID]
	if len(queued) == 0 {
		return &PullOfflineResponse{Messages: []Message{}}, nil
	}

	limit := request.Limit
	if limit <= 0 || limit > len(queued) {
		limit = len(queued)
	}

	messages := make([]Message, limit)
	copy(messages, queued[:limit])

	remaining := make([]Message, len(queued)-limit)
	copy(remaining, queued[limit:])
	service.offlineByUser[request.UserID] = remaining

	return &PullOfflineResponse{Messages: messages}, nil
}

func main() {
	address := flag.String("addr", "127.0.0.1:19100", "message_query rpc server listen address")
	flag.Parse()

	server := rpcserver.NewServer()
	if err := server.RegisterService(NewMessageService()); err != nil {
		log.Fatalf("register MessageService failed: %v", err)
	}

	fmt.Printf("message_query server listening on %s\n", *address)
	if err := server.ListenAndServe("tcp", *address); err != nil {
		log.Fatalf("message_query server stopped with error: %v", err)
	}
}
