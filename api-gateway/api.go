package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	services []ServiceRegistrationReq
	wg       sync.WaitGroup
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
	mux.HandleFunc("/register", makeHttpHandleFunc(register))

	log.Printf("Listening to port %s\n", server.Address)

	go monitorServices()

	http.ListenAndServe(server.Address, mux)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Gateway is up"))
}

func register(w http.ResponseWriter, r *http.Request) error {
	var request ServiceRegistrationReq
	err := json.NewDecoder(r.Body).Decode(&request)

	defer r.Body.Close()

	if err != nil {
		return writeJson(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
	}

	services = append(services, request)
	log.Printf("%+v\n", services)
	return writeJson(w, map[string]string{"message": "Service registered to the gateway..."}, http.StatusOK)
}

func writeJson(w http.ResponseWriter, data any, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func makeHttpHandleFunc(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJson(w, map[string]string{"error": err.Error()}, http.StatusOK)
		}
	}
}

func monitorServices() {
	for {
		if len(services) == 0 || services == nil {
			log.Println("No Services are registered yet...")
		} else {
			for _, req := range services {
				wg.Add(1)
				go fetchHealthStatus(req, &wg)
			}
			wg.Wait()
		}
		time.Sleep(time.Minute * 1)
	}
}

func fetchHealthStatus(req ServiceRegistrationReq, wg *sync.WaitGroup) {
	response, err := http.Get(req.BasePath /*+ "/health"*/)
	if err != nil {
		log.Printf("%s service unresponsive due to %s\n", strings.ToUpper(req.Name), err)
	} else {
		log.Printf("Recieved %d from %s service\n", response.StatusCode, strings.ToUpper(req.Name))
	}
	wg.Done()
}
