package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type TxnReq struct {
	Type string  `json:"type"`
	Txn  [][]any `json:"txn"`
}

type TxnRes struct {
	Type string  `json:"type"`
	Txn  [][]any `json:"txn"`
}

type SyncMsg struct {
	Type  string      `json:"type"`
	State map[int]int `json:"state"`
}

func main() {
	node := maelstrom.NewNode()
	store := make(map[int]int)
	var mu sync.RWMutex

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		for range ticker.C {
			mu.RLock()
			snapshot := make(map[int]int)

			for k, v := range store {
				snapshot[k] = v
			}
			mu.RUnlock()
			msg := SyncMsg{
				Type:  "sync",
				State: snapshot,
			}
			for _, peer := range node.NodeIDs() {
				if peer != node.ID() {
					node.Send(peer, msg)
				}
			}
		}
	}()

	node.Handle("sync", func(msg maelstrom.Message) error {
		var req SyncMsg
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}
		mu.Lock()
		defer mu.Unlock()
		for k, v := range req.State {
			if currVal, exists := store[k]; !exists || v > currVal {
				store[k] = v
			}
		}
		return nil
	})

	node.Handle("txn", func(msg maelstrom.Message) error {
		var req TxnReq

		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}
		mu.Lock()
		for _, txn := range req.Txn {
			op, key := txn[0].(string), int(txn[1].(float64))
			switch op {
			case "r":
				if val, exists := store[key]; exists {
					txn[2] = val
				} else {
					txn[2] = nil
				}
			case "w":
				store[key] = int(txn[2].(float64))
			}
		}
		mu.Unlock()
		res := TxnRes{
			Type: "txn_ok",
			Txn:  req.Txn,
		}
		return node.Reply(msg, res)
	})

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}
