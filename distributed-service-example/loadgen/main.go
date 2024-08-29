package main

import (
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {
	frontend := os.Getenv("FRONTEND")
	if frontend == "" {
		frontend = "http://localhost:8080"
	}

	waitTime, err := time.ParseDuration(os.Getenv("WAIT_TIME"))
	if err != nil {
		waitTime = 0
		if os.Getenv("WAIT_TIME") != "" {
			log.Printf("can't parse WAIT_TIME: %s. Assuming zero", err.Error())
		}
	}

	avgMean := float64(25000)
	meanDev := float64(20000)
	dev := float64(5000)
	for {
		log.Println("requesting index")
		resp, err := http.Get(frontend + "/")
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}
		_, _ = io.ReadAll(resp.Body)

		mean := 5 + avgMean + math.Sin(float64(time.Now().Unix())*math.Pi/(20*60))*meanDev
		num := getInt(int(mean-dev), int(mean+dev))
		log.Printf("(mean: %f) Sending form with number %d", mean, num)
		start := time.Now()
		resp, err = http.PostForm(frontend+"/submit", url.Values{"number": []string{strconv.Itoa(num)}})
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}
		_, _ = io.ReadAll(resp.Body)
		log.Printf("%d took %d ms", resp.StatusCode, time.Now().Sub(start).Milliseconds())
		if waitTime > 0 {
			log.Printf("waiting %s before the next invocation", waitTime)
			time.Sleep(waitTime)
		}
	}
}

func getInt(min, max int) int {
	mean := float64(min) + float64(max)/2
	dev := float64(max) - mean
	for {
		candidate := int(rand.NormFloat64()*dev + mean)
		if candidate >= min && candidate <= max {
			return candidate
		}
	}
}
