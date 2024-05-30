package services

import (
	"fmt"
	"net/http"
)

func StartServer() {
	fmt.Println("Starting the HTTP Server ...")
	mux := http.NewServeMux()
	mux.HandleFunc("/app", ApplicationHandler)
	mux.HandleFunc("/validate", ValidationHandler)
	mux.HandleFunc("/wd/hub/", SessionHandler)
	mux.HandleFunc("/", GlobalHandler)
	handler := middleware(mux)
	http.ListenAndServe(":4723", handler)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		basicAuthToken := r.Header.Get("Authorization")
		if basicAuthToken == "" || !isValidUser(basicAuthToken) {
			http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
