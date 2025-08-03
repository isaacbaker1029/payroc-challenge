// main
package main

import (
	"time"
)

// Main helper method
func main() {
	// Config Items
	backendServers := []string{
		"localhost:9001",
		"localhost:9002",
		"localhost:9003",
	}
	listenerPort := ":8080"
	healthCheckInterval := 10 * time.Second
	// ---------

	loadBalancer := NewLoadBalancer(listenerPort, backendServers)

	startHealthChecks(loadBalancer, healthCheckInterval)

	loadBalancer.Start()
}
