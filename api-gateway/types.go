package main

import "net/http"

type ServiceRegistrationReq struct {
	Name     string `json:"name"`
	BasePath string `json:"base_path"`
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error
