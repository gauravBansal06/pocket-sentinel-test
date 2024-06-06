package services

import (
	"byod/common"
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"syscall"
)

var srv *http.Server

const isHttpsEnabled = false

// certificate and key as strings
const certPEM = `-----BEGIN CERTIFICATE-----
MIIDOTCCAiGgAwIBAgIUclXDcAXIdUDO1ckmjk8cK7kPADYwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCSU4xEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yNDA1MjEwOTQ2NTZaFw0yNTA1
MjEwOTQ2NTZaMEUxCzAJBgNVBAYTAklOMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQDUfjtm3iOcWtCbMGX+5wBUhKhXg8Wa8TremqeKvb6C
gmX7pk+xyX9Fb8DZpL7GwS822d8ps2J7Ya23L4Qrq+ofyxYvEBobAP91SR3FdOym
oCcTX+t05wuaJMUCrnDo7E1DA5cJUYoF8bVuDfCOhpBVkb6da5EkJVRXOKNZM7cW
EaQr8JJVYLWPPUOIhauqYsvUfw3uYwiDyWzeVGwoW7t9DK51/Krqchn3ziPSWJvq
zpWWpqUr1opnlbMfWdYcL8L7lwRSFb8W25qhSNcN7XpDYBtitkwG3SvMoSjVV817
X8KWWlV1S2g1WC7BEMyr10plI+PgUB7SnfudQy+a2KO/AgMBAAGjITAfMB0GA1Ud
DgQWBBTmpcBsCzyW2IHxyihbtjqR4ZkorzANBgkqhkiG9w0BAQsFAAOCAQEAas19
qixs1darRrF6Nv7DVRqac18Ls6u1gixjzHK37+fUYHMFjUGOKObafxWixN4VC3Oj
xP4jPqajTJ3L7HPqnZrJc2LjE6e8f9SoWsa8G/jeJa8Qdq1M2lPoeSAj4yFVok8F
b3EzBOcJpifDyjMzhKLE9p0WMGJoe3SqbpbetY9IMBP+yt8c9MHMPwqb1bgOEeBp
m5Dzr8in7rE68b3F8K3pDqonfykz9hea7n5VCAtrHoSx60wTiBjyo8DDfhgpsZWg
OPuh9qy5T8c4u/UciEzM52wLq5vbSS5XY+7XLwNC4//FZevrcTYJesNZlrxVgdJr
pgdEFxJgiNsQqe5mlg==
-----END CERTIFICATE-----`

const keyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDUfjtm3iOcWtCb
MGX+5wBUhKhXg8Wa8TremqeKvb6CgmX7pk+xyX9Fb8DZpL7GwS822d8ps2J7Ya23
L4Qrq+ofyxYvEBobAP91SR3FdOymoCcTX+t05wuaJMUCrnDo7E1DA5cJUYoF8bVu
DfCOhpBVkb6da5EkJVRXOKNZM7cWEaQr8JJVYLWPPUOIhauqYsvUfw3uYwiDyWze
VGwoW7t9DK51/Krqchn3ziPSWJvqzpWWpqUr1opnlbMfWdYcL8L7lwRSFb8W25qh
SNcN7XpDYBtitkwG3SvMoSjVV817X8KWWlV1S2g1WC7BEMyr10plI+PgUB7Snfud
Qy+a2KO/AgMBAAECggEAMnf2yNRa1eB16l4A451e5TQKvZpk/QttZYCK6Xduf5XT
/gR+qiXG7imAPtETpKYufhMaH0/lRJOrDDajaDHxLfLwxmXHJYHDqsoeYY8HQ/4G
ueHOzRmxFj/EcQIYfCHfqdp47XQp2YaShlyWNWu9bS1r/fyV4OVi1kJz/ZQ8WGeS
cJXiDRQT33LgUfiy2A0v6mFXoGZjwmoaztJqrVBtMQl19UXKAcfCUSlvApj5zYrh
CvxqQoLM1T78g04Btp1VLJzMIpxzhMI/ShIFBgogOw6OpqC/Nct+d2z/gAPSDqPv
9VisoI/NPpfrgUldcwD2B7P0IUJR+xrJHqBZlfWxXQKBgQDrD2x4TpOCFyC01au0
WHh7w/fIHXuz9tUHi6PobuaLBqOa0r61KsY52K2rhZAjlalt6EFIk3/B0DqG5ZHr
CZwE8pRdjPvQWWjDtH6GtKnSGMli6ru5KSvIXcEgBmTOC+bLmJ6Ea3/4afF35op6
LZ27X+N2r2jbtTusAawEHZVGVQKBgQDnbClnjGMfADx3uabceZeeoQY1KwXNoker
bY9CbXKuXLGS0uZ5yhHLK32WgO+B/IkyzSbeYoZPBFnWGSA2MMfVD7Z7qBaIwl63
YAbYGSLsdT5VPjyCyemZQ6VQHM99A6o6o8KcBMeVSU5BqwEZxVM5G921WNx8VXZw
HbErI4rNwwKBgQCtjWHFIh7/OhxH6EsyiO/MUdCszDf9lA2N2KhYgSlvFKPPODLe
iIp7Q2RW8/KMk3/ZSlaJQ35cl2XfG7k1FI9Bh+nLeMCkAJ+9f5K72sBYAz0N78pS
1/cfpTlmb9IV3+uz/ydrFgQSYgaLRIiR1QRUyGOlybVeIt3ADiv3jfAdUQKBgENn
LWgLa5NL7lCwsfjlMVPycmxp63bJHTbA4kjmt9AxD0dERfyS7jvOnvWG+DtT4mH5
fqim6Zd6HPBOwSMHciyMNwotGuMaOZwPS+8E4zcbrtwFFHrDdFY/bZa3zXcL6jjK
GZR3j+nbP//AXsGyx1qK0zhOeWl9OtHM1B1MbNEhAoGBAJt14S5pyZg0n8DsXT7u
HGIrgjqS5maRVJpdn8msIggfxng+aaYoMcbHn2zRSEyRdekUPtUU8PNSSbp5H7j9
0DTId/fMKpbg6Pj11j/LT5BXmvhVq9SRIluw0R/Y0ljUb0eiF6s0Uiv/BSQfIQSc
klXUfNYOUhOJ2KqSQPzNjz+M
-----END PRIVATE KEY-----`

// StartServer initializes and starts an HTTP server on a specified port.
func StartServer() {
	mux := http.NewServeMux()
	setupRoutes(mux)

	if isHttpsEnabled {
		StartHttpsServer(mux)
		return
	}

	log.Println("Starting the HTTP Server on port 4723...")

	srv = &http.Server{
		Addr:    ":4723",
		Handler: middleware(mux),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("error starting http server: ", err)
			log.Println("Please make sure that port 4723 is free")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}()
}

func StartHttpsServer(mux *http.ServeMux) {
	log.Println("Starting the HTTPS Server on port 4723...")

	// Create the TLS certificate
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		log.Println("error starting server: ", err)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		return
	}

	// Create a TLS config with the certificate
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	srv = &http.Server{
		Addr:      ":4723",
		Handler:   middleware(mux),
		TLSConfig: tlsConfig,
	}

	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Println("error starting https server: ", err)
			log.Println("Please make sure that port 4723 is free")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
	}()
}

func KillServer(ctx context.Context) {
	log.Println("shutting down server...")
	if srv == nil {
		log.Println("server already not started")
		return
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown error: ", err)
	}
	log.Println("server stopped..")
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
		if userInfo, ok := authenticateRequest(r); ok {
			// Set user info in context for futher use and authorizations
			ctx := context.WithValue(r.Context(), common.UserContextKey, userInfo)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, `{"status":"Unauthorized"}`, http.StatusUnauthorized)
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
func authenticateRequest(r *http.Request) (common.UserDetails, bool) {
	authToken := r.Header.Get("Authorization")
	if authToken == "" {
		return common.UserDetails{}, false
	}
	return IsValidUser(authToken)
}
