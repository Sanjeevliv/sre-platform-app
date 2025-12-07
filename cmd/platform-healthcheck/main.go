package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: healthcheck <url>")
		os.Exit(1)
	}

	url := os.Args[1]
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error checking %s: %v\n", url, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unhealthy status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	os.Exit(0)
}
