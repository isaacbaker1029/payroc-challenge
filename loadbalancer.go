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
	listenerPort       string
	allBackends        []string
	healthyBackends    []string
	mu                 sync.RWMutex // RWMutex is now used for better read performance
	backendConnections map[string]int
}

// NewLoadBalancer creates a new LoadBalancer instance
func NewLoadBalancer(port string, backends []string) *LoadBalancer {
	return &LoadBalancer{
		listenerPort:       port,
		allBackends:        backends,
		healthyBackends:    make([]string, 0), // Start with no healthy backends until watchdog runs
		backendConnections: make(map[string]int),
	}
}

// getNextBackend selects the next healthy backend using round-robin
func (lb *LoadBalancer) getNextBackend() (string, error) {
	lb.mu.RLock() // Use a read-lock, allowing multiple connections at once
	defer lb.mu.RUnlock()

	if len(lb.healthyBackends) == 0 {
		return "", fmt.Errorf("no healthy backends available")
	}

	// Find the backend with the minimum number of connections
	minConnections := -1
	var bestBackend string
	for _, backendAddr := range lb.healthyBackends {
		count := lb.backendConnections[backendAddr]
		if minConnections == -1 || count < minConnections {
			minConnections = count
			bestBackend = backendAddr
		}
	}
	return bestBackend, nil
}

// handleConnection manages an incoming client connection
func (lb *LoadBalancer) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	backendAddr, err := lb.getNextBackend()
	if err != nil {
		log.Printf("Failed to get a backend for %s: %v", clientConn.RemoteAddr(), err)
		return
	}

	// Increment connection count for the chosen backend
	lb.mu.Lock()
	lb.backendConnections[backendAddr]++
	lb.mu.Unlock()

	// Decrement the connection count when the handler exits
	defer func() {
		lb.mu.Lock()
		lb.backendConnections[backendAddr]--
		lb.mu.Unlock()
	}()

	log.Printf("Forwarding connection from %s to %s (current connections: %d)", clientConn.RemoteAddr(), backendAddr, lb.backendConnections[backendAddr])

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

	// Prune the connection counts for any backends that are no longer healthy
	currentCounts := make(map[string]int)
	for _, backendAddr := range newHealthyBackends {
		// Carry over the count if the backend was already known
		if count, ok := lb.backendConnections[backendAddr]; ok {
			currentCounts[backendAddr] = count
		} else {
			currentCounts[backendAddr] = 0
		}
	}

	lb.healthyBackends = newHealthyBackends
	lb.backendConnections = currentCounts
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
