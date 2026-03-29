# Gossip Glomers Challenge

Welcome to my implementation of the Gossip Glomers Challenge! This project is a collection of distributed systems exercises inspired by the Maelstrom framework from Fly.io. My goal here is to deeply understand distributed systems concepts by building them from scratch, not just reading about them.

## Why I Built This

I've always been fascinated by distributed systems—their complexity, the subtle bugs, and the elegance of their solutions. Instead of just reading papers or watching lectures, I wanted to get my hands dirty. The Maelstrom challenges are a perfect playground for this: they force you to think about real-world problems like message delivery, consistency, and fault tolerance.

## Project Structure

This repo contains several sub-projects, each corresponding to a Maelstrom challenge:

- **maelstrom-echo**: The classic echo server, a warm-up for message passing.
- **maelstrom-broadcast**: Broadcast and gossip protocols for reliable message delivery.
- **maelstrom-counter**: Distributed counters and dealing with eventual consistency.
- **maelstrom-unique-ids**: Generating unique IDs in a distributed system.
- **maelstrom-kafka**: Building a simple distributed log (like Kafka).
- **maelstrom-txn**: Handling distributed transactions and isolation.

Each subfolder is a standalone Go module, so you can work on them independently.

## How to Run

Each challenge is a Go program. To run any of them:

1. `cd` into the desired subproject (e.g., `maelstrom-echo`).
2. Run `go build` to build the binary.
3. Use the Maelstrom tool to run the challenge (see the [Fly.io Maelstrom docs](https://github.com/jepsen-io/maelstrom) for details).

Example:
```sh
cd maelstrom-echo
go build
/path/to/maelstrom test -w echo --bin ./maelstrom-echo --node-count 1 --time-limit 10
```

## My Approach & Thinking

I approached each challenge with a focus on clarity and correctness first, then performance. I tried to:
- Write code that's easy to read and reason about.
- Prefer explicitness over cleverness—distributed bugs are hard enough already.
- Learn from failures: if something didn't work, I left notes or TODOs.

## What I Learned

- The importance of message ordering and idempotency.
- How subtle network issues can break naive implementations.
- That distributed systems are hard, but also really fun when you see them work!
---

*If you have feedback, suggestions, or want to discuss distributed systems, feel free to reach out or open an issue!*
