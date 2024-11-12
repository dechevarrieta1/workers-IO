package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

const (
	numRequests = 1000
	numWorkers  = 10
)

func simulateTask() {
	time.Sleep(time.Duration(rand.Intn(10)+1) * time.Millisecond)
}

func runWithoutOptimizations(requests int) time.Duration {
	start := time.Now()
	for i := 0; i < requests; i++ {
		simulateTask()
	}
	return time.Since(start)
}

func runWithWorkerPool(requests, workers int) time.Duration {
	start := time.Now()
	var wg sync.WaitGroup
	taskChan := make(chan struct{}, requests)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range taskChan {
				simulateTask()
			}
		}()
	}

	for i := 0; i < requests; i++ {
		taskChan <- struct{}{}
	}
	close(taskChan)

	wg.Wait()
	return time.Since(start)
}

func generateLatencyPlot(latencyWithout, latencyWith []int) error {
	p := plot.New()

	p.Title.Text = "Latencia en un Sistema Con y Sin Worker Pool y Actor Model"
	p.X.Label.Text = "Número de solicitudes"
	p.Y.Label.Text = "Latencia (ms)"

	withoutPts := make(plotter.XYs, len(latencyWithout))
	withPts := make(plotter.XYs, len(latencyWith))
	for i := 0; i < len(latencyWithout); i++ {
		x := float64((i + 1) * 100)
		withoutPts[i].X = x
		withoutPts[i].Y = float64(latencyWithout[i])
		withPts[i].X = x
		withPts[i].Y = float64(latencyWith[i])
	}

	err := plotutil.AddLinePoints(p,
		"Sin Worker Pool y Actor Model", withoutPts,
		"Con Worker Pool y Actor Model", withPts)
	if err != nil {
		return err
	}

	if err := p.Save(8*vg.Inch, 4*vg.Inch, "latency_comparison.png"); err != nil {
		return err
	}

	fmt.Println("Gráfico guardado como 'latency_comparison.png'")
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	latencyWithoutOptimizations := []int{}
	latencyWithOptimizations := []int{}

	for requests := 100; requests <= numRequests; requests += 100 {
		latency1 := runWithoutOptimizations(requests).Milliseconds()
		latencyWithoutOptimizations = append(latencyWithoutOptimizations, int(latency1))

		latency2 := runWithWorkerPool(requests, numWorkers).Milliseconds()
		latencyWithOptimizations = append(latencyWithOptimizations, int(latency2))

		fmt.Printf("Latencia sin optimización para %d solicitudes: %d ms\n", requests, latency1)
		fmt.Printf("Latencia con optimización para %d solicitudes: %d ms\n", requests, latency2)
	}

	err := generateLatencyPlot(latencyWithoutOptimizations, latencyWithOptimizations)
	if err != nil {
		fmt.Println("Error generando el gráfico:", err)
		os.Exit(1)
	}
}
