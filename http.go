package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func newModuleHandler(mp ModuleProvider) http.Handler {
	handler := mux.NewRouter()
	handler.HandleFunc("/.well-known/terraform.json", discover)
	handler.HandleFunc("/mod/{namespace}/{name}/{provider}/versions", version(mp))
	handler.HandleFunc("/mod/{namespace}/{name}/{provider}/{version}/download", download)
	handler.HandleFunc("/mod/{namespace}/{name}/{provider}/{version}/blob.tar.gz", blob(mp))
	return handler
}

func httpErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func logger(next http.Handler) http.Handler {
	logger := handlers.CombinedLoggingHandler(os.Stdout, next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		debug := false
		if debug {
			req, _ := httputil.DumpRequest(r, false)
			log.Println(string(req))
		}
		logger.ServeHTTP(w, r)
	})
}

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}

		// do auth here

		next.ServeHTTP(w, r)
	})
}

func discover(w http.ResponseWriter, r *http.Request) {
	discoveryResponse := struct {
		ModulePath string `json:"modules.v1"`
	}{
		ModulePath: "/mod",
	}

	w.Header().Add("Content-Type", "application/json")
	resp, err := json.Marshal(discoveryResponse)
	if err != nil {
		log.Print("failed to marshal discover response", err)
		httpErr(w, err)
		return
	}
	w.Write(resp)
}

func getModule(r *http.Request) Module {
	vars := mux.Vars(r)
	return Module{
		Namespace: vars["namespace"],
		Name:      vars["name"],
		Provider:  vars["provider"],
	}
}

type moduleVersionResponse struct {
	Modules []moduleVersions `json:"modules"`
}

type moduleVersions struct {
	Versions []moduleVersion `json:"versions"`
}

type moduleVersion struct {
	Version string `json:"version"`
}

func versionsToResponse(versions []string) *moduleVersionResponse {
	resp := &moduleVersionResponse{
		Modules: []moduleVersions{
			{},
		},
	}
	for _, v := range versions {
		resp.Modules[0].Versions = append(resp.Modules[0].Versions, moduleVersion{Version: v})
	}
	return resp
}

func version(mp ModuleProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		module := getModule(r)

		versions, err := mp.GetVersions(r.Context(), module)
		if err != nil {
			log.Print("failed to get versions", err)
			httpErr(w, err)
			return
		}

		resp, err := json.Marshal(versionsToResponse(versions))
		if err != nil {
			log.Print("failed to marshal versions response", err)
			httpErr(w, err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(resp)
	}
}

func download(w http.ResponseWriter, r *http.Request) {
	m := getModule(r)
	version := mux.Vars(r)["version"]
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("X-Terraform-Get", fmt.Sprintf("/mod/%s/%s/%s/%s/blob.tar.gz", m.Namespace, m.Name, m.Provider, version))
}

func blob(mp ModuleProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := getModule(r)
		version := mux.Vars(r)["version"]

		reader, err := mp.GetReader(r.Context(), m, version)
		if err != nil {
			log.Print("failed to get blob reader", err)
			httpErr(w, err)
			return
		}

		_, err = io.Copy(w, reader)
		if err != nil {
			log.Print("failed to deliver blob", err)
			return
		}
	}
}
