package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
    _RES = "../res"
)

func handleFile(w http.ResponseWriter, r *http.Request, ftype string) {
	file, err := os.Open(_RES + r.URL.Path)
	if err != nil {
		fmt.Fprintf(w, "%v", err)
	} else {
		w.Header().Add("content-type", ftype)
		io.Copy(w, file)
		file.Close()
	}
}

func fixPath(r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = "/index.html"
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fixPath(r)
    handleFile(w, r, MimeType(r.URL.Path))
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println(http.ListenAndServe(":80", nil))
}
