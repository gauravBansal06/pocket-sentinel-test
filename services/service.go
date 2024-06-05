package services

import (
	"context"
	"log"
	"net/http"
	"syscall"
)

var server *http.Server

// StartServer initializes and starts an HTTP server on a specified port.
func StartServer() {
	log.Println("Starting the HTTP Server on port 4723...")
	mux := http.NewServeMux()
	setupRoutes(mux)
	server = &http.Server{
		Addr:    ":4723",
		Handler: middleware(mux),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("error starting server")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}()
}

func KillServer() {
	log.Println("shutting down server")
	if server == nil {
		log.Println("server already not started")
		return
	}
	if err := server.Shutdown(context.Background()); err != nil {
		log.Println("Server Shutdown error: ", err)
	}
}

// setupRoutes configures the URL endpoints and their corresponding handlers.
func setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/app", ApplicationHandler)     // Handle application-specific actions
	mux.HandleFunc("/validate", ValidationHandler) // Handle validation actions
	mux.HandleFunc("/wd/hub/", SessionHandler)     // Handle WebDriver sessions
	mux.HandleFunc("/", GlobalHandler)             // Handle all other requests
}

// middleware applies various HTTP headers and controls the request flow.
func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		setCORSHeaders(w)

		// Allow preflight checks for CORS
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Authenticate the request
		if authenticateRequest(r) {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})
}

// setCORSHeaders sets the necessary CORS headers for each request.
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// authenticateRequest checks if the provided request is authorized.
func authenticateRequest(r *http.Request) bool {
	basicAuthToken := r.Header.Get("Authorization")
	return basicAuthToken != "" && IsValidUser(basicAuthToken)
}
