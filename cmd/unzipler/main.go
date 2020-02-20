package main

import (
	"encoding/json"
	"log"
	"net/http"
	"unzipler/internal/utils"
	"unzipler/pkg/arch"
)

type Error struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", index)
	router.HandleFunc("/handleArchive", handleArchive)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}

func index(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte{})
}

func handleArchive(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fpath, fname, err := utils.UploadArchive(r)
		if err != nil {
			log.Fatal(err)
		}
		a := arch.New(fpath, fname)
		tree, err := a.Unzip()
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(tree)
	default:
		notFound(w)
	}
}

func notFound(w http.ResponseWriter) {
	error(w, "404 page not found", http.StatusNotFound)
}

func error(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Error{
		Error: error,
		Code:  code,
	})
}
