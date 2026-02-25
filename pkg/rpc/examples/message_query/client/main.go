package main

import (
	"context"
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
	serviceAddress := "127.0.0.1:19100"

	client, err := rpcclient.NewClient(
		serviceAddress,
		rpcclient.WithTimeout(3*time.Second),
		rpcclient.WithRetryMax(2),
		rpcclient.WithRetryBackoff(20*time.Millisecond, 200*time.Millisecond, 0.1),
		rpcclient.WithIdempotentMethod("MessageService.SendMessage"),
		rpcclient.WithIdempotentMethod("MessageService.PullOffline"),
	)
	if err != nil {
		log.Fatalf("new client failed: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	sendRequest := &SendMessageRequest{
		FromUserID:  "userA",
		ToUserID:    "userB",
		Content:     "hello from userA",
		ClientMsgID: fmt.Sprintf("userA-%d", time.Now().UnixNano()),
	}
	var sendResponse SendMessageResponse
	if err = client.Call(context.Background(), "MessageService.SendMessage", sendRequest, &sendResponse); err != nil {
		log.Fatalf("send message failed: %v", err)
	}
	fmt.Printf("send ack: message_id=%d duplicated=%v\n", sendResponse.AckMessageID, sendResponse.Duplicated)

	pullRequest := &PullOfflineRequest{UserID: "userB", Limit: 10}
	var pullResponse PullOfflineResponse
	if err = client.Call(context.Background(), "MessageService.PullOffline", pullRequest, &pullResponse); err != nil {
		log.Fatalf("pull offline failed: %v", err)
	}

	fmt.Printf("pull offline for userB, count=%d\n", len(pullResponse.Messages))
	for index, message := range pullResponse.Messages {
		fmt.Printf("[%d] id=%d from=%s to=%s content=%q client_msg_id=%s sent_at=%d\n", index, message.MessageID, message.FromUserID, message.ToUserID, message.Content, message.ClientMsgID, message.SentAtUnix)
	}
}
