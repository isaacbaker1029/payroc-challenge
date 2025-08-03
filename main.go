// main
package main

// Main helper method
func main() {
	backendServers := []string{
		"localhost:9001",
		"localhost:9002",
		"localhost:9003",
	}
	listenerPort := ":8080"

	loadBalancer := NewLoadBalancer(listenerPort, backendServers)
	loadBalancer.Start()
}
