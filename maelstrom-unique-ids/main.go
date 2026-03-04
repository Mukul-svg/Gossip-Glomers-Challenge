package main

import (
	"log"

	"github.com/google/uuid"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

// Define the exact shape of your response
type GenerateResponse struct {
	Type string    `json:"type"`
	ID   uuid.UUID `json:"id"`
}

func main() {
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		// Construct the strictly-typed response
		resp := GenerateResponse{
			Type: "generate_ok",
			ID:   uuid.New(),
		}

		// n.Reply automatically handles linking the message IDs
		return n.Reply(msg, resp)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
