package main

import (
	"net"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	// deploy handler should follow the path
	mux.HandleFunc("/deploy/", deployHandler)

	nl, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}
	err = http.Serve(nl, mux)
	if err != nil {
		panic(err)
	}
}
