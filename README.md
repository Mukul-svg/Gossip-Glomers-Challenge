# Gossip Glomers

I've read a lot about distributed systems.Designing data intensive applications book, various papers, the usual suspects. And for a while I convinced myself that reading was enough—that I understood consensus, idempotency, network partitions. Then I actually tried to *build* something, and found out pretty quickly that I didn't.

This repo is my attempt to fix that. It's my solutions to [Fly.io's Gossip Glomers challenges](https://github.com/jepsen-io/maelstrom), a series of distributed systems exercises built on top of Maelstrom (and underneath that, Jepsen). Each challenge gives you a problem—broadcast messages across a cluster, generate globally unique IDs, implement a distributed transaction—and then Maelstrom simulates the nastiest network conditions it can think of and tells you whether your system held up.

Spoiler: mine often didn't. At first.

---

## The Challenges

Each subfolder is a standalone Go module. They build on each other conceptually, even if the code doesn't.

- **maelstrom-echo** — A warm-up. You receive a message, you send it back. Sounds trivial, and it is, but it forces you to actually read the wire protocol and get your message-handling scaffolding right before things get complicated.

- **maelstrom-unique-ids** — Generate unique IDs across multiple nodes with zero coordination. My first instinct was to reach for a shared counter. That's wrong—coordination is exactly what you can't rely on. I ended up concatenating a nanosecond timestamp with a cryptographically random integer. `math/rand` wasn't enough; the test actually caught collisions.

- **maelstrom-broadcast** — This is where it gets real. The task sounds simple: make sure every node eventually learns about every message. My first implementation was about 15 lines and failed immediately. The issues: no retry logic when messages got dropped, no idempotency checks so re-delivered messages corrupted state, and a naive flood topology that sent way too many messages once the efficiency constraints kicked in. Getting broadcast right—fault-tolerant *and* efficient—took longer than all the other challenges combined.

- **maelstrom-counter** — A grow-only distributed counter, built on top of Maelstrom's sequentially consistent key-value store. The trap here is thinking you can just increment. You can't. Concurrent increments race. The insight is that each node tracks its *own* count, and the total is computed by summing all of them.

- **maelstrom-kafka** — Building a minimal distributed log. Append-only, committed offsets, multiple consumers. The ordering guarantees matter a lot here and force you to think carefully about what "committed" actually means.

- **maelstrom-txn** — Distributed transactions with read-write isolation. The hardest one. This is where I started leaving comments like `// this is probably wrong` and `// TODO: understand why this works`.

---

## How to Run

Each challenge is a self-contained Go program.

```sh
cd maelstrom-broadcast
go build
/path/to/maelstrom test -w broadcast --bin ./maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10
```

See the [Maelstrom docs](https://github.com/jepsen-io/maelstrom) for the full test parameters per challenge. Fair warning: when something fails, Maelstrom's error output is a serialized Clojure data structure. It takes a minute to learn to read it.

---

## What Actually Surprised Me

**Idempotency isn't optional.** In any system where you retry messages—which you have to, because networks drop packets—your message handlers need to be safe to call more than once. I knew this conceptually. I didn't internalize it until I watched my broadcast state silently corrupt itself on duplicate deliveries.

**"At-least-once" is only useful if you handle duplicates.** These two ideas are a package deal, and the second one is easy to forget.

**Topology matters more than I expected.** In the broadcast challenge, a naive "send to everyone" approach passes the correctness tests but blows past the efficiency constraints—you need to send fewer than 20 messages per operation across 25 nodes. I experimented with tree topologies, which reduced messages dramatically but created hot-spots. If the root node goes down, the whole thing breaks. There's a real tension between efficiency and resilience that no amount of reading had made concrete to me.

**Maelstrom finds bugs that normal testing never would.** It injects partitions, delays specific nodes, drops messages probabilistically. The bugs it surfaces aren't the kind you'd catch in a unit test or even a local integration test. That's kind of the point.

---

## Approach

I tried to solve each challenge in the dumbest correct way first, then make it smarter only when the tests demanded it. This turned out to be the right call—several times I wrote something clever that was also subtly wrong, and falling back to something obvious fixed it.

Go felt like a natural fit here. The goroutine model makes concurrent message handling readable, and the Maelstrom Go library handles the wire protocol so you can focus on the actual logic. Though I'd be curious how this feels in Rust, where the type system would probably catch a few of the races I found at runtime.

---

## Still Left to Do

- The `maelstrom-txn` solution works, but I'm not confident I fully understand *why* under all partitioning scenarios. Needs more thought.
- I want to go back to broadcast and benchmark the topology changes more carefully—right now I've found something that passes, but I don't have a strong intuition for why one topology beats another at specific node counts.

---

If you've worked through these challenges and want to compare notes, feel free to open an issue. I'm especially curious how others handled the efficiency constraints in broadcast without sacrificing fault tolerance.
