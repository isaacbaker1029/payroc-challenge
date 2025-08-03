// loadbalancer
package main

import (
	"fmt"
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

// Start is a stub for now — just shows it's being called
func (lb *LoadBalancer) Start() {
	fmt.Printf("Starting load balancer on %s with %d backends...\n", lb.listenerPort, len(lb.backends))
}
