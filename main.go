package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type SimpleServer struct {
	Address string
	Proxy   *httputil.ReverseProxy
}

func (s SimpleServer) AddressString() string {
	return s.Address
}

func (s SimpleServer) IsAlive() bool {
	return true
}

func (s SimpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.Proxy.ServeHTTP(rw, req)

}

type Server interface {
	AddressString() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, request *http.Request)
}

type LoadBalancer struct {
	Port            string
	RoundRobinCount int
	Servers         []Server
}

func NewServer(addr string) *SimpleServer {

	serverUrl, err := url.Parse(addr)

	handleErr(err)

	server := SimpleServer{
		Address: addr,
		Proxy:   httputil.NewSingleHostReverseProxy(serverUrl),
	}

	return &server

}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	loadBalancer := LoadBalancer{
		Port:            port,
		Servers:         servers,
		RoundRobinCount: 0,
	}
	return &loadBalancer
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("Error %v\n", err)
		os.Exit(1)
	}
}

func (lb *LoadBalancer) getNextAvailableFunction() Server {

	server := lb.Servers[lb.RoundRobinCount%len(lb.Servers)]
	for !server.IsAlive() {
		lb.RoundRobinCount++
		server = lb.Servers[lb.RoundRobinCount%len(lb.Servers)]
	}
	lb.RoundRobinCount++

	return server

}

func (lb *LoadBalancer) serverProxy(rw http.ResponseWriter, r *http.Request) {

	targetServer := lb.getNextAvailableFunction()

	fmt.Printf("Forwarding request to server %q\n", targetServer.AddressString())

	targetServer.Serve(rw, r)
}

func main() {

	//Creating array of servers which will handle requests
	servers := []Server{
		NewServer("http://google.com"),
		NewServer("https://bing.com"),
		NewServer("http://duckduckgo.com"),
	}

	// A new Load balancer which will forward all the request it receives to one of the servers
	// based on round robin count and server's availablity
	lb := NewLoadBalancer("8080", servers)

	HandleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serverProxy(rw, req)
	}

	http.HandleFunc("/", HandleRedirect)

	fmt.Printf("Serving request at `localhost:%s`\n", lb.Port)
	http.ListenAndServe(":"+lb.Port, nil)

}
