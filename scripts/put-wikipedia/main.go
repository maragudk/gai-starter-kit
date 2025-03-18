package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	numWorkers  = 8
	maxRetries  = 3
	apiEndpoint = "http://localhost:8080/documents"
)

func main() {
	// Parse command line flags
	inputDir := flag.String("input", "pages", "Input directory containing documents to upload")
	flag.Parse()

	// Check if directory exists
	if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
		fmt.Printf("Error: Input directory %s does not exist\n", *inputDir)
		os.Exit(1)
	}

	// Set up channels for workers
	jobs := make(chan string, 10000)
	var wg sync.WaitGroup

	// Start worker pool
	for range numWorkers {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	// Track statistics
	var fileCount int

	// Walk the directory and queue files
	fmt.Println("Scanning for files...")
	err := filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process markdown files
		if filepath.Ext(path) == ".md" {
			jobs <- path
			fileCount++
		}

		return nil
	})
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d files to process\n", fileCount)

	// Close the jobs channel to signal workers that no more jobs are coming
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()

	fmt.Printf("Upload complete: processed %d files\n", fileCount)
}

// worker processes jobs from the queue
func worker(jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range jobs {
		// Process the file with retries
		success := false
		for attempt := range maxRetries {
			if attempt > 0 {
				// Wait before retry
				backoff := time.Duration(500*(1<<uint(attempt))) * time.Millisecond
				jitter := time.Duration(rand.IntN(250)) * time.Millisecond
				time.Sleep(backoff + jitter)
			}

			if err := uploadFile(path); err != nil {
				fmt.Printf("Error uploading %s (attempt %d/%d): %v\n", path, attempt+1, maxRetries, err)
			} else {
				success = true
				break
			}
		}

		if !success {
			fmt.Printf("Failed to upload after %d attempts: %s\n", maxRetries, path)
		}
	}
}

// uploadFile reads a file and uploads it to the API
func uploadFile(path string) error {
	// Read file content
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	defer file.Close()

	// Create the request
	req, err := http.NewRequest(http.MethodPost, apiEndpoint, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/markdown")

	// Send the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf("unexpected status code: %d (%v)", resp.StatusCode, string(body))
	}

	return nil
}
