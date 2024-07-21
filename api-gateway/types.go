package main

import "net/http"

type ServiceRegistrationReq struct {
	Name      string   `json:"name"`
	BasePath  string   `json:"base_path"`
	Endpoints []string `json:"endpoints"`
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error

type UserData struct {
	Name     string `json:"name"`
	LastName string `json:"lastname"`
	Gender   string `json:"gender"`
}
