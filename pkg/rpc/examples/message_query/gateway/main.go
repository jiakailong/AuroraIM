package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	rpcclient "kama_chat_server/pkg/rpc/client"
)

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

type Message struct {
	MessageID   uint64 `json:"message_id"`
	FromUserID  string `json:"from_user_id"`
	ToUserID    string `json:"to_user_id"`
	Content     string `json:"content"`
	ClientMsgID string `json:"client_msg_id"`
	SentAtUnix  int64  `json:"sent_at_unix"`
}

type PullOfflineResponse struct {
	Messages []Message `json:"messages"`
}

func main() {
	address := flag.String("addr", "127.0.0.1:19100", "message service rpc address")
	action := flag.String("action", "send", "send|pull")
	fromUserID := flag.String("from", "userA", "sender user id")
	toUserID := flag.String("to", "userB", "receiver user id")
	content := flag.String("content", "hello from gateway", "message content")
	clientMsgID := flag.String("client_msg_id", fmt.Sprintf("cli-%d", time.Now().UnixNano()), "client side message id")
	pullUserID := flag.String("user", "userB", "pull offline messages for user")
	limit := flag.Int("limit", 20, "pull message limit")
	flag.Parse()

	client, err := rpcclient.NewClient(
		*address,
		rpcclient.WithTimeout(3*time.Second),
		rpcclient.WithRetryMax(2),
		rpcclient.WithRetryBackoff(20*time.Millisecond, 200*time.Millisecond, 0.1),
		rpcclient.WithIdempotentMethod("MessageService.SendMessage"),
		rpcclient.WithIdempotentMethod("MessageService.PullOffline"),
	)
	if err != nil {
		log.Fatalf("new rpc client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	switch *action {
	case "send":
		request := &SendMessageRequest{
			FromUserID:  *fromUserID,
			ToUserID:    *toUserID,
			Content:     *content,
			ClientMsgID: *clientMsgID,
		}
		var response SendMessageResponse
		if err = client.Call(context.Background(), "MessageService.SendMessage", request, &response); err != nil {
			log.Fatalf("gateway send failed: %v", err)
		}
		fmt.Printf("gateway send ack: message_id=%d duplicated=%v\n", response.AckMessageID, response.Duplicated)
	case "pull":
		request := &PullOfflineRequest{UserID: *pullUserID, Limit: *limit}
		var response PullOfflineResponse
		if err = client.Call(context.Background(), "MessageService.PullOffline", request, &response); err != nil {
			log.Fatalf("gateway pull failed: %v", err)
		}
		fmt.Printf("gateway pull result for %s: count=%d\n", *pullUserID, len(response.Messages))
		for index, message := range response.Messages {
			fmt.Printf("[%d] id=%d from=%s to=%s content=%q client_msg_id=%s sent_at=%d\n", index, message.MessageID, message.FromUserID, message.ToUserID, message.Content, message.ClientMsgID, message.SentAtUnix)
		}
	default:
		log.Fatalf("unsupported action: %s", *action)
	}
}
