// watchdog
package main

import (
	"log"
	"net"
	"time"
)

// startHealthChecks launches a background goroutine to periodically check backend health.
func startHealthChecks(lb *LoadBalancer, interval time.Duration) {
	// Run the first check immediately to populate the healthy list
	performHealthCheck(lb)

	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			performHealthCheck(lb)
		}
	}()
}

// performHealthCheck dials all potential backends and updates the load balancer's healthy list.
func performHealthCheck(lb *LoadBalancer) {
	var currentlyHealthy []string
	for _, backendAddr := range lb.allBackends {
		conn, err := net.DialTimeout("tcp", backendAddr, 2*time.Second)
		if err == nil {
			conn.Close()
			currentlyHealthy = append(currentlyHealthy, backendAddr)
		}
	}
	log.Printf("Health check complete. Healthy backends: %d/%d", len(currentlyHealthy), len(lb.allBackends))
	lb.updateHealthyBackends(currentlyHealthy)
}
