package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/crypto_pickle/bench/express"
	"github.com/guptarohit/asciigraph"
)

var filepath *string = flag.String("log", "default.txt", "directory to save bench mark results")

func BenchmarkEndpoint(regex string, duration time.Duration, workers int, record_duration time.Duration) {
	var req_count, req_err, req_time int64
	stop := make(chan struct{})

	fmt.Printf("Generating Unique Endpoints for %s... \n", regex)

	exp := express.Compile(regex)
	urls := express.GenerateUnique(exp, 100000)

	for i := 0; i < 5; i++ {
		fmt.Printf(" - %s \n", urls[i])
	}

	fmt.Println("Continue? (y/n)")

	var con string
	fmt.Scanln(&con)

	if con == "n" {
		return
	}

	recorder, record := time.NewTicker(record_duration), make([]float64, int(duration/record_duration))
	go func() {
		for i := 0; i < len(record); i++ {
			t, c := float64(atomic.LoadInt64(&req_time)), float64(atomic.LoadInt64(&req_count))
			if c == 0 {
				record[i] = t / 1000
			} else {
				record[i] = t / c / 1000
			}
			<-recorder.C
		}
	}()

	ticker, wg := time.Tick(duration), sync.WaitGroup{}
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-stop:
					return
				default:
					t1 := time.Now()

					resp, err := http.Get(urls[rand.Intn(len(urls))])
					if err != nil {
						atomic.AddInt64(&req_err, 1)
					} else {
						wg.Add(1)
						go func() {
							defer wg.Done()

							_, err := io.ReadAll(resp.Body)
							resp.Body.Close()

							if err != nil {
								atomic.AddInt64(&req_err, 1)
							}
						}()
					}
					t2 := time.Now()

					atomic.AddInt64(&req_count, 1)
					atomic.AddInt64(&req_time, t2.UnixMilli()-t1.UnixMilli())
				}
			}
		}()
	}

	<-ticker
	for i := 0; i < workers; i++ {
		stop <- struct{}{}
	}

	wg.Wait()

	rqs := float64(req_count) / float64(req_time) * 1000
	latency := duration.Milliseconds() / req_count

	log.Printf("Sent %d requests over %.2f s. Got \n Average Latency: %d ms \n Average RPS: %.4f \n Errors: %d \n", req_count, duration.Seconds(), latency, rqs, req_err)

	graph := asciigraph.Plot(record)
	log.Printf("Graph: \n %s \n", graph)
}

func main() {
	const (
		duration     time.Duration = time.Minute * 10
		workers      int           = 2
		record_timer               = time.Second * 10
	)

	startLogger()

	regex := "http://localhost:8080/get-orderbooks\\?symbol=(btc)|(eth)|(bnb)|(bch)|(xrp)|(ltc)|(ftm)|(arb)|(xvg)|(sol)usdt&start=2023-07-(0[8-9])#dT((0[1-9])|(1[0-9])|(2[0-4]))#h:((0[1-9])|([1-5][0-9]))#m:00.0&end=2023-07-#dT#h:#m:(0[0-9])|(1[0-9]).[0-9](&depth=([1-9][0-9][0-9])|([1-4][0-9][0-9][0-9])|(5000))?(&freq=(1)|(10))?"
	// regex := "http://localhost:8080/get-orderbooks\\?symbol=(btc)|(eth)|(bnb)|(bch)|(xrp)|(ltc)|(ftm)|(arb)|(xvg)|(sol)usdt&start=2023-07-08T11:49:00.0&end=2023-07-08T11:49:(0[0-9])|(1[0-9]).[0-9](&depth=([1-9][0-9][0-9])|([1-4][0-9][0-9][0-9])|(5000))?"

	BenchmarkEndpoint(regex, duration, workers, record_timer)
}

func startLogger() {
	flag.Parse()

	file, err := os.OpenFile("bench/runs/"+*filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
	log.Println("New Logger Started")
}
