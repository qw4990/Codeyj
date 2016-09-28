package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
)

var (
	getHTML  string
	postHTML string
)

func init() {
	buf, err := ioutil.ReadFile("./templates/get.html")
	if err != nil {
		panic(err)
	}
	getHTML = string(buf)

	buf, err = ioutil.ReadFile("./templates/post.html")
	if err != nil {
		panic(err)
	}
	postHTML = string(buf)
}

func main() {
	http.HandleFunc("/fucking", handler)
	fmt.Println(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html")

	if r.Method == "GET" {
		doGET(w, r)
	} else if r.Method == "POST" {
		doPOST(w, r)
	}
}

func doGET(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	word := r.Form.Get("word")
	id, words := GenSentence(word)

	mp := map[string]interface{}{
		"words": words,
		"id":    id,
	}

	t := template.Must(template.New("xxx").Parse(getHTML))
	t.Execute(w, mp)
}

func doPOST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	words := make([]string, 0, 30)
	idstr := r.Form.Get("id")
	id, _ := strconv.Atoi(idstr)

	for i := 0; i < len(GetSentence(id)); i++ {
		key := fmt.Sprintf("word%v", i)
		val := r.Form.Get(key)
		words = append(words, val)
	}

	result, newWords := CheckSentence(id, words)

	mp := map[string]interface{}{
		"words":  newWords,
		"result": result,
	}

	t := template.Must(template.New("xxx").Parse(postHTML))
	t.Execute(w, mp)
}
