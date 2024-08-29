package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	Port    int    `env:"PORT" envDefault:"8080"`
	Backend string `env:"BACKEND"`
}

//go:embed html/index.html
var index []byte

const result = `
<html><body>
<a href="/">Go back to index</a>
<hr/>
<div style="word-break: break-all;">
%s! = <br/>
%s
<hr/>
</div>
<a href="/">Go back to index</a>
</body></html>
`

func main() {
	cfg := Config{}
	panicOnErr(env.Parse(&cfg))
	mux := http.NewServeMux()
	mux.HandleFunc("/submit", func(rw http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		input := req.FormValue("number")
		// will simulate an internal error from time to time
		nmbr, _ := strconv.Atoi(input)
		if rand.Intn(int(math.Max(10.0, float64(nmbr/10)))) == 0 {
			log.Println("returning random error")
			rw.WriteHeader(500 + rand.Intn(5))
			return
		}

		resp, err := http.Get(cfg.Backend + "/factorial/" + input)
		if err != nil {
			log.Println("GET error ", err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			return
		}
		if body, err := io.ReadAll(resp.Body); err != nil {
			log.Println("Read error ", err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
		} else {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.WriteHeader(resp.StatusCode)
			rw.Write([]byte(fmt.Sprintf(result, input, string(body))))
		}
	})
	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.Write(index)
	})
	panicOnErr(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux))
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
