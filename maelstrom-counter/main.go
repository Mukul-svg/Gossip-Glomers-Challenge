package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type AddReq struct {
	Type  string `json:"type"`
	Delta int    `json:"delta"`
}
type AddRes struct {
	Type string `json:"type"`
}

type ReadReq struct {
	Type string `json:"type"`
}
type ReadRes struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

func main() {
	node := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(node)
	key := "delta"

	node.Handle("add", func(msg maelstrom.Message) error {
		var req AddReq

		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}
		for {
			curr_delta, new_delta := 0, 0
			curr_delta, err := kv.ReadInt(context.Background(), key)
			if err != nil {
				var rpcErr *maelstrom.RPCError

				if errors.As(err, &rpcErr) && rpcErr.Code == maelstrom.KeyDoesNotExist {
					curr_delta = 0
				} else {
					return err
				}
			}

			new_delta = curr_delta + req.Delta

			err = kv.CompareAndSwap(context.Background(), key, curr_delta, new_delta, true)
			if err == nil {
				break
			}
		}
		res := AddRes{
			Type: "add_ok",
		}
		return node.Reply(msg, res)
	})

	node.Handle("read", func(msg maelstrom.Message) error {
		var value int

		value, err := kv.ReadInt(context.Background(), key)
		if err != nil {
			var rpcErr *maelstrom.RPCError

			if errors.As(err, &rpcErr) && rpcErr.Code == maelstrom.KeyDoesNotExist {
				value = 0
			} else {
				return err
			}
		}

		res := ReadRes{
			Type:  "read_ok",
			Value: value,
		}
		return node.Reply(msg, res)
	})

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}
