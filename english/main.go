package main

import (
	"fmt"
    "bufio"
	"io"
	"net/http"
	"os"
    "math/rand"
)

const (
    _RES = "../res"

    _HTML_FORM = `<form action="" method="post">
    <textarea name="sentence" rows="3" cols="60">
</textarea>
  <input type="submit" value="Submit" />
</form>`
    _HTML_HEAD = `<html><head>
    <title>nyadb</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    </head>`
    _HTML_TAIL = `</html>`

)

var (
    _SENTENCES_ENGLISH []string
    _SENTENCES_CHINESE []string
    _SENTENCES_TOT int
    _RAND_NUMBER int
)

func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("content-type", "text/html")

    fmt.Fprint(w, _HTML_HEAD)
    defer fmt.Fprint(w, _HTML_TAIL)

    if r.Method == "POST" {
        r.ParseForm()
        sentence := r.Form["sentence"][0]
        if sentence == _SENTENCES_ENGLISH[_RAND_NUMBER] {
            fmt.Fprint(w, "<p><font color='green'>Right</font></p>")
            _RAND_NUMBER = rand.Int() % _SENTENCES_TOT
        } else {
            fmt.Fprintf(w, "<p><font color='red'>Wrong: %s</font></p>", _SENTENCES_ENGLISH[_RAND_NUMBER])
        }
    }

    fmt.Fprint(w, "<p>"+_SENTENCES_CHINESE[_RAND_NUMBER]+"</p>")
    fmt.Fprint(w, _HTML_FORM)
}

func initSentences() {
    file, err := os.Open("sentences.txt")
    if err != nil {
        panic(err)
    }
    reader := bufio.NewReader(file)

    _SENTENCES_TOT = 0
    eFlag := true
    for {
        line, err := reader.ReadString('\n')
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }
        line = line[:len(line)-1]
        if len(line) == 0 { // blank line
            continue
        } else if line[0] == '#' { // commend line
            continue
        }

        if eFlag == true {
            _SENTENCES_ENGLISH = append(_SENTENCES_ENGLISH, line)
        } else {
            _SENTENCES_CHINESE = append(_SENTENCES_CHINESE, line)
            _SENTENCES_TOT++
        }
        eFlag = !eFlag
    }
}

func main() {
    initSentences()
	http.HandleFunc("/fucking", handler)
	fmt.Println(http.ListenAndServe(":8080", nil))
}