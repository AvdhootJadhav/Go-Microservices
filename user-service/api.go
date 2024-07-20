package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	endpoints = make([]string, 0)
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
	mux.HandleFunc("/user", makeHttpHandleFunc(getUser))

	err := mux.Walk(gorillaWalkIn)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Listening to port", server.Address)
	http.ListenAndServe(server.Address, mux)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("User Service is up..."))
}

func getUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return writeJson(w, map[string]string{"name": "Avdhoot", "lastname": "J", "gender": "M"}, 200)
	}
	return writeJson(w, map[string]string{"error": "method not allowed"}, http.StatusBadRequest)
}

func makeHttpHandleFunc(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusOK)
		}
	}
}

func writeJson(w http.ResponseWriter, data any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func gorillaWalkIn(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	path, err := route.GetPathTemplate()
	if err == nil {
		endpoints = append(endpoints, path)
	}
	return err
}
