package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	services     = make(map[string]ServiceRegistrationReq)
	wg           sync.WaitGroup
	healthStatus = make(map[string]bool) // record health status of service
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
	mux.HandleFunc("/user", makeHttpHandleFunc(fetchUser))
	mux.HandleFunc("/monitor", makeHttpHandleFunc(fetchHealthData))

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

	services[request.Name] = request
	return writeJson(w, map[string]string{"message": "Service registered to the gateway..."}, http.StatusOK)
}

func fetchUser(w http.ResponseWriter, r *http.Request) error {
	for _, data := range services {
		for _, paths := range data.Endpoints {
			if paths == r.URL.Path {
				response, err := http.Get(data.BasePath + paths)
				if err != nil {
					return err
				}
				defer response.Body.Close()
				userData := UserData{}
				json.NewDecoder(response.Body).Decode(&userData)
				return writeJson(w, userData, http.StatusOK)
			}
		}
	}
	return nil
}

func fetchHealthData(w http.ResponseWriter, r *http.Request) error {
	if len(healthStatus) == 0 {
		return fmt.Errorf("No data found")
	}
	return writeJson(w, healthStatus, http.StatusOK)
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
			for name, data := range services {
				wg.Add(1)
				go fetchHealthStatus(name, data.BasePath, &wg)
			}
			wg.Wait()
		}
		time.Sleep(time.Minute * 1)
	}
}

func fetchHealthStatus(name, path string, wg *sync.WaitGroup) {
	log.Printf("Checking Health for %s Service\n", name)
	_, err := http.Get(path + "/health")
	if err != nil {
		healthStatus[name] = false
	} else {
		healthStatus[name] = true
	}
	wg.Done()
}
