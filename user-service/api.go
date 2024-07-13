package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	Address string
}

func NewServer(address string) APIServer {
	return APIServer{
		Address: address,
	}
}

func (server APIServer) Run() {
	mux := mux.NewRouter()
	mux.HandleFunc("/health", healthCheck)

	log.Println("Listening to port", server.Address)
	http.ListenAndServe(server.Address, mux)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("User Service is up..."))
}
