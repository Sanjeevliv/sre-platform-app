package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	url := "http://localhost:8080/healthz"
	totalRequests := 200 // Default limit is 100 RPS, so this should trigger some 429s

	fmt.Println("Starting rate limit test...")
	start := time.Now()

	statusCodes := make(map[int]int)
	var mu sync.Mutex

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(url)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			statusCodes[resp.StatusCode]++
			mu.Unlock()
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("Completed %d requests in %v\n", totalRequests, duration)
	for code, count := range statusCodes {
		fmt.Printf("Status %d: %d\n", code, count)
	}
}
