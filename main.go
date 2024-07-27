package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

const (
	numWorkers = 10
)

type Request struct {
	id int
}

type Result struct {
	id       int
	duration time.Duration
}

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func heavyComputation() {
	// Calcular números primos en un rango grande para simular carga pesada
	for i := 0; i < 10000; i++ {
		isPrime(rand.Intn(100000))
	}
}

func worker(id int, requests <-chan Request, results chan<- Result) {
	for req := range requests {
		start := time.Now()
		fmt.Printf("Worker %d processing request %d\n", id, req.id)
		heavyComputation() // Simula trabajo pesado
		duration := time.Since(start)
		results <- Result{id: req.id, duration: duration}
	}
}

func startWorkerPool(numWorkers int, requests <-chan Request, results chan<- Result) {
	for w := 1; w <= numWorkers; w++ {
		go worker(w, requests, results)
	}
}

func createPlot(durations []time.Duration) error {
	pts := make(plotter.XYs, len(durations))
	for i, duration := range durations {
		pts[i].X = float64(i)
		pts[i].Y = duration.Seconds()
	}

	p := plot.New()

	p.Title.Text = "Request Processing Times"
	p.X.Label.Text = "Request ID"
	p.Y.Label.Text = "Time (seconds)"

	line, err := plotter.NewLine(pts)
	if err != nil {
		return err
	}

	p.Add(line)

	if err := p.Save(8*vg.Inch, 4*vg.Inch, "processing_times.png"); err != nil {
		return err
	}
	return nil
}

func main() {
	requests := make(chan Request, 200)
	results := make(chan Result, 200)

	startWorkerPool(numWorkers, requests, results)

	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		numRequestsStr := query.Get("num")
		numRequests, err := strconv.Atoi(numRequestsStr)
		if err != nil {
			http.Error(w, "Invalid number of requests", http.StatusBadRequest)
			return
		}

		durations := make([]time.Duration, numRequests)

		for r := 1; r <= numRequests; r++ {
			requests <- Request{id: r}
		}

		for a := 1; a <= numRequests; a++ {
			result := <-results
			durations[result.id-1] = result.duration
		}

		if err := createPlot(durations); err != nil {
			http.Error(w, "Failed to create plot", http.StatusInternalServerError)
			log.Println("Failed to create plot:", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Processed %d requests. See processing_times.png for the plot.\n", numRequests)
	})

	log.Fatal(http.ListenAndServe(":8082", nil))
}
