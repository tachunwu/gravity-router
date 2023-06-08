package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type Client struct {
	nc *nats.Conn
}

func NewClient() *Client {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	return &Client{nc: nc}
}

func (c *Client) Set(ctx context.Context, key string, value []byte, replicas int) error {
	inbox := nats.NewInbox()
	// Listen for replies on the inbox
	sub, err := c.nc.SubscribeSync(inbox)
	if err != nil {
		return fmt.Errorf("subscribe error: %w", err)
	}
	defer sub.Unsubscribe()

	msg := &nats.Msg{
		Subject: key,
		Data:    value,
		Header:  make(nats.Header),
		Reply:   inbox,
	}
	msg.Header.Add("op", "set")

	err = c.nc.PublishMsg(msg)
	if err != nil {
		return fmt.Errorf("publish error: %w", err)
	}

	for i := 0; i < replicas; i++ {
		_, err := sub.NextMsgWithContext(ctx)
		if err != nil {
			return fmt.Errorf("receive reply error: %w", err)
		}
	}

	return nil
}

func (c *Client) Del(ctx context.Context, key string, replicas int) error {
	inbox := nats.NewInbox()
	// Listen for replies on the inbox
	sub, err := c.nc.SubscribeSync(inbox)
	if err != nil {
		return fmt.Errorf("subscribe error: %w", err)
	}
	defer sub.Unsubscribe()

	msg := &nats.Msg{
		Subject: key,
		Header:  make(nats.Header),
		Reply:   inbox,
	}
	msg.Header.Add("op", "del")

	err = c.nc.PublishMsg(msg)
	if err != nil {
		return fmt.Errorf("publish error: %w", err)
	}

	for i := 0; i < replicas; i++ {
		_, err := sub.NextMsgWithContext(ctx)
		if err != nil {
			return fmt.Errorf("receive reply error: %w", err)
		}
	}

	return nil
}

func (c *Client) Get(key string) {
	inbox := nats.NewInbox()
	// Listen for replies on the inbox
	sub, err := c.nc.SubscribeSync(inbox)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	msg := &nats.Msg{
		Subject: key,
		Header:  make(nats.Header),
		Reply:   inbox,
	}
	msg.Header.Add("op", "get")

	err = c.nc.PublishMsg(msg)
	if err != nil {
		log.Fatal(err)
	}

	reply, err := sub.NextMsg(1 * time.Second)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get operation reply:", reply)
}

func main() {

	client := NewClient()

	key := "key"
	value := []byte("value")

	err := client.Set(context.Background(), key, value, 3)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Set operation completed")

	client.Get(key)

	err = client.Del(context.Background(), key, 3)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Del operation completed")

}
