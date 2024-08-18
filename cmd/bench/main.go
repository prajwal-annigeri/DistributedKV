package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const N = 100

var addr = flag.String("addr", "localhost:8080", "HTTP address")
var iterations = flag.Int("iterations", 10000, "No. of iterations")
var readIterations = flag.Int("read-iterations", 10000, "No. of read iterations")
var concurrency = flag.Int("concurrency", 1, "No. of goroutines in parallel")

func benchmark(name string, iterations int, fn func() string) (float64, []string) {
	var max time.Duration
	min := time.Hour

	var strs []string

	start := time.Now()
	for i := 0; i < iterations; i++ {
		iterStart := time.Now()
		strs = append(strs, fn())
		iterDuration := time.Since(iterStart)
		if iterDuration > max {
			max = iterDuration
		}
		if iterDuration < min {
			min = iterDuration
		}
	}

	avg := time.Since(start) / time.Duration(iterations)
	qps := float64(iterations) / (float64(time.Since(start)) / float64(time.Second))
	fmt.Printf("Func %s took average: %s, %.1f qps, maximum: %s, minimum: %s\n", name, avg, qps, max, min)

	return qps, strs
}

func writeRandomKey() string {
	key := fmt.Sprintf("key-%d", rand.Intn(100000))
	value := fmt.Sprintf("value-%d", rand.Intn(100000))

	values := url.Values{
		key:   []string{key},
		value: []string{value},
	}

	response, err := http.Get("http://" + *addr + "/set" + values.Encode())
	if err != nil {
		log.Fatalf("Error during set key: %v\n", err)
	}
	defer response.Body.Close()

	return key
}

func readRandomKey(allKeys []string) string {
	key := allKeys[rand.Intn(len(allKeys))]

	values := url.Values{
		"key": []string{key},
	}

	response, err := http.Get("http://" + *addr + "/get" + values.Encode())
	if err != nil {
		log.Fatalf("Error during get key: %v\n", err)
	}
	defer response.Body.Close()

	return key
}

func benchmarkWrite() []string {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalQps float64
	var allKeys []string

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			qps, strs := benchmark("write", *iterations, writeRandomKey)
			mu.Lock()
			totalQps += qps
			allKeys = append(allKeys, strs...)
			mu.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()

	log.Printf("Write total QPS: %.1f, set %d keys\n", totalQps, len(allKeys))

	return allKeys
}

func benchmarkRead(allKeys []string) {
	var totalQps float64
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			qps, _ := benchmark("read", *readIterations, func() string { return readRandomKey(allKeys) })
			mu.Lock()
			totalQps += qps
			mu.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()

	log.Printf("Read total QPS: %.1f\n", totalQps)
}

func main() {
	flag.Parse()

	allKeys := benchmarkWrite()
	benchmarkRead(allKeys)
	
}
