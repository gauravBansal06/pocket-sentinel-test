package services

import (
	"net/http"
)

func GlobalHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`OK`))
}
