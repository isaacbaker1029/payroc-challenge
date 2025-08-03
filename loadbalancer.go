// loadbalancer
package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// LoadBalancer holds the state for our load balancer
// sync Mutex is used similar to 'synchronized' in java, makes parts of the application single threaded to avoid contentions
type LoadBalancer struct {
	listenerPort string
	backends     []string
	mu           sync.Mutex
	nextBackend  int
}

// NewLoadBalancer creates a new LoadBalancer instance
func NewLoadBalancer(port string, backends []string) *LoadBalancer {
	return &LoadBalancer{
		listenerPort: port,
		backends:     backends,
		nextBackend:  0,
	}
}

// removeBackend safely removes a backend from the pool
func (lb *LoadBalancer) removeBackend(addr string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	var updatedBackends []string
	for _, backendAddr := range lb.backends {
		if backendAddr != addr {
			updatedBackends = append(updatedBackends, backendAddr)
		}
	}

	if len(updatedBackends) < len(lb.backends) {
		lb.backends = updatedBackends
		log.Printf("Removed backend %s from the pool. %d backends remaining.", addr, len(lb.backends))
		if lb.nextBackend >= len(lb.backends) {
			lb.nextBackend = 0
		}
	}
}

// getNextBackend selects the next available backend using round-robin
func (lb *LoadBalancer) getNextBackend() (string, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.backends) == 0 {
		return "", fmt.Errorf("no available backends")
	}

	backend := lb.backends[lb.nextBackend]
	lb.nextBackend = (lb.nextBackend + 1) % len(lb.backends)
	return backend, nil
}

// handleConnection manages an incoming client connection
func (lb *LoadBalancer) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	backendAddr, err := lb.getNextBackend()
	if err != nil {
		log.Printf("Failed to get a backend for %s: %v", clientConn.RemoteAddr(), err)
		return
	}

	log.Printf("Forwarding connection from %s to %s", clientConn.RemoteAddr(), backendAddr)

	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Printf("Failed to connect to backend %s: %v", backendAddr, err)
		lb.removeBackend(backendAddr)
		return
	}
	defer backendConn.Close()

	go io.Copy(backendConn, clientConn)
	io.Copy(clientConn, backendConn)
}

// Start begins the load balancer's listening loop
func (lb *LoadBalancer) Start() {
	listener, err := net.Listen("tcp", lb.listenerPort)
	if err != nil {
		log.Fatalf("Failed to start listener on port %s: %v", lb.listenerPort, err)
	}
	defer listener.Close()
	log.Printf("Load balancer listening on %s", lb.listenerPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept new connection: %v", err)
			continue
		}
		go lb.handleConnection(conn)
	}
}
