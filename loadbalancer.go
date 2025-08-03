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
	listenerPort    string
	allBackends     []string
	healthyBackends []string
	mu              sync.RWMutex // RWMutex is now used for better read performance
	nextBackend     int
}

// NewLoadBalancer creates a new LoadBalancer instance
func NewLoadBalancer(port string, backends []string) *LoadBalancer {
	return &LoadBalancer{
		listenerPort:    port,
		allBackends:     backends,
		healthyBackends: make([]string, 0), // Start with no healthy backends until watchdog runs
		nextBackend:     0,
	}
}

// getNextBackend selects the next healthy backend using round-robin
func (lb *LoadBalancer) getNextBackend() (string, error) {
	lb.mu.RLock() // Use a read-lock, allowing multiple connections at once
	defer lb.mu.RUnlock()

	if len(lb.healthyBackends) == 0 {
		return "", fmt.Errorf("no healthy backends available")
	}

	backend := lb.healthyBackends[lb.nextBackend]
	lb.nextBackend = (lb.nextBackend + 1) % len(lb.healthyBackends)
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
		log.Printf("Failed to connect to healthy backend %s: %v", backendAddr, err)
		return
	}
	defer backendConn.Close()

	go io.Copy(backendConn, clientConn)
	io.Copy(clientConn, backendConn)
}

// updateHealthyBackends is called by the watchdog to update the list of healthy servers.
func (lb *LoadBalancer) updateHealthyBackends(newHealthyBackends []string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.healthyBackends = newHealthyBackends
	// Reset index if it's out of bounds after the update
	if lb.nextBackend >= len(lb.healthyBackends) {
		lb.nextBackend = 0
	}
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
