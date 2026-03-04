package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type SendReq struct {
	Type string `json:"type"`
	Key  string `json:"key"`
	Msg  int    `json:"msg"`
}

type SendRes struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
}

type PollReq struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

type PollRes struct {
	Type string              `json:"type"`
	Msgs map[string][][2]int `json:"msgs"`
}

type CommitReq struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

type CommitRes struct {
	Type string `json:"type"`
}

type ListReq struct {
	Type string   `json:"type"`
	Keys []string `json:"keys"`
}

type ListRes struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

func main() {
	node := maelstrom.NewNode()
	kv := maelstrom.NewLinKV(node)

	node.Handle("send", func(msg maelstrom.Message) error {
		var req SendReq
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}

		topic := req.Key
		msg_id := req.Msg

		for {
			var current_log [][2]int

			raw_val, err := kv.Read(context.Background(), topic)
			if err != nil {
				var rpcErr *maelstrom.RPCError
				if errors.As(err, &rpcErr) && rpcErr.Code == maelstrom.KeyDoesNotExist {
					current_log = make([][2]int, 0)
				} else {
					return err
				}
			} else {
				b, _ := json.Marshal(raw_val)
				json.Unmarshal(b, &current_log)
			}

			my_offset := len(current_log)

			new_log := append(current_log, [2]int{my_offset, msg_id})

			err = kv.CompareAndSwap(context.Background(), topic, raw_val, new_log, true)
			if err == nil {
				res := SendRes{
					Type:   "send_ok",
					Offset: my_offset,
				}
				return node.Reply(msg, res)
			}
		}
	})

	node.Handle("poll", func(msg maelstrom.Message) error {
		var req PollReq
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}

		poll_response := make(map[string][][2]int)

		for topic, req_offset := range req.Offsets {
			var current_log [][2]int

			raw_val, err := kv.Read(context.Background(), topic)
			if err == nil {
				b, _ := json.Marshal(raw_val)
				json.Unmarshal(b, &current_log)

				for _, msg_pair := range current_log {
					if msg_pair[0] >= req_offset {
						poll_response[topic] = append(poll_response[topic], msg_pair)
					}
				}
			}
		}

		res := PollRes{
			Type: "poll_ok",
			Msgs: poll_response,
		}
		return node.Reply(msg, res)
	})

	node.Handle("commit_offsets", func(msg maelstrom.Message) error {
		var req CommitReq
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}

		for topic, offset := range req.Offsets {
			commit_key := "commit_" + topic

			kv.Write(context.Background(), commit_key, offset)
		}

		res := CommitRes{
			Type: "commit_offsets_ok",
		}
		return node.Reply(msg, res)
	})

	node.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		var req ListReq
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}

		res_offsets := make(map[string]int)

		for _, topic := range req.Keys {
			commit_key := "commit_" + topic

			offset, err := kv.ReadInt(context.Background(), commit_key)
			if err == nil {
				res_offsets[topic] = offset
			}
		}

		res := ListRes{
			Type:    "list_committed_offsets_ok",
			Offsets: res_offsets,
		}
		return node.Reply(msg, res)
	})

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}
}
