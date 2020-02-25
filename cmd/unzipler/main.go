package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	router.HandleFunc("/files/", handleFiles)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}

func index(w http.ResponseWriter, _ *http.Request) {
	notFound(w)
}

func handleFiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		p := strings.TrimPrefix(r.URL.Path, "/files/")
		p = "./assets/files/" + p

		if _, err := os.Stat(p); os.IsNotExist(err) {
			notFound(w)
			return
		}

		streamBytes, err := ioutil.ReadFile(p)

		if err != nil {
			log.Fatal(err)
		}

		b := bytes.NewBuffer(streamBytes)

		ctype := http.DetectContentType(streamBytes[:512])
		w.Header().Set("Content-type", ctype)
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if _, err := b.WriteTo(w); err != nil {
			log.Fatal(err)
		}
	default:
		notAllowed(w)
	}
}

func handleArchive(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fpath, fname, err := utils.UploadArchive(r)
		if err != nil {
			log.Fatal(err)
		}
		fname = strings.TrimSuffix(fname, filepath.Ext(fname))
		a := arch.New(fpath, fname)
		tree, err := a.Unzip()
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(tree)
	default:
		notAllowed(w)
	}
}

func notFound(w http.ResponseWriter) {
	error(w, "404 page not found", http.StatusNotFound)
}

func notAllowed(w http.ResponseWriter) {
	error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
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
