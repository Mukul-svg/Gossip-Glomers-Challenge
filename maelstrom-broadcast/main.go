package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type BroadcastRequest struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

type BroadcastResponse struct {
	Type string `json:"type"`
}

type ReadResponse struct {
	Type     string `json:"type"`
	Messages []int  `json:"messages"`
}

type TopologyResponse struct {
	Type string `json:"type"`
}

type TopologyRequest struct {
	Type     string              `json:"type"`
	Topology map[string][]string `json:"topology"`
}

func main() {
	n := maelstrom.NewNode()
	var message_list []int
	var neighbors []string
	var mu sync.RWMutex

	n.Handle("broadcast", func(msg maelstrom.Message) error {

		var req BroadcastRequest

		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}

		mu.Lock()
		alreadySeen := false
		for _, v := range message_list {
			if v == req.Message {
				alreadySeen = true
				break
			}
		}
		if !alreadySeen {
			message_list = append(message_list, req.Message)
		}
		mu.Unlock()

		if alreadySeen {
			return n.Reply(msg, BroadcastResponse{
				Type: "broadcast_ok",
			})
		}

		for _, neighbor := range neighbors {
			if msg.Src != neighbor {
				go func(dest string, msgInt int) {
					for {
						ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
						_, err := n.SyncRPC(ctx, dest, BroadcastRequest{
							Type:    "broadcast",
							Message: msgInt,
						})
						cancel()

						if err == nil {
							break
						}
					}
				}(neighbor, req.Message)
			}
		}

		resp := BroadcastResponse{
			Type: "broadcast_ok",
		}

		return n.Reply(msg, resp)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		mu.Lock()
		resp := ReadResponse{
			Type:     "read_ok",
			Messages: message_list,
		}
		mu.Unlock()
		return n.Reply(msg, resp)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var req TopologyRequest

		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}

		neighbors = req.Topology[n.ID()]

		resp := TopologyResponse{
			Type: "topology_ok",
		}

		return n.Reply(msg, resp)
	})

	n.Handle("broadcast_ok", func(msg maelstrom.Message) error {
		return nil
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
