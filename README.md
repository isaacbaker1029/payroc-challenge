# Payroc Load Balancer Challenge

Bit of a quick project to build a simple Layer 4 (TCP) load balancer in Go.

## What it does

- Listens for incoming TCP connections (on port 8080)
- Forwards them to a list of backend servers (e.g. :9001, :9002)
- Uses round robin to balance the load
- If a backend is down, it just skips it
- Handles multiple clients using goroutines

## How to run it

### Start some fake backends (e.g. using `ncat`)
```bash
ncat -lk 9001
ncat -lk 9002
