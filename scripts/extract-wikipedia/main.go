package main

import (
	"crypto/sha256"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// XML structures to match Wikipedia dump format
type MediaWiki struct {
	XMLName xml.Name `xml:"mediawiki"`
	Pages   []Page   `xml:"page"`
}

type Page struct {
	Title    string    `xml:"title"`
	NS       int       `xml:"ns"`
	ID       int       `xml:"id"`
	Redirect *Redirect `xml:"redirect"`
	Revision Revision  `xml:"revision"`
}

type Redirect struct {
	Title string `xml:"title,attr"`
}

type Revision struct {
	Text Text `xml:"text"`
}

type Text struct {
	Content string `xml:",chardata"`
}

// Worker pool configuration
const (
	numWorkers     = 8     // Number of worker goroutines
	filesPerDir    = 10000 // Maximum files per directory
	maxQueueLength = 100   // Maximum length of the job queue
)

// Job represents a page to be processed
type Job struct {
	Title string
	Text  string
	ID    int
}

func main() {
	// Parse command-line flags
	inputFile := flag.String("input", "", "Path to Wikipedia XML dump file")
	outputDir := flag.String("output", "pages", "Output directory for extracted pages")
	limit := flag.Int("limit", 0, "Maximum number of pages to process (0 = no limit)")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: Input file is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	err := os.MkdirAll(*outputDir, 0755)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Open XML file
	file, err := os.Open(*inputFile)
	if err != nil {
		fmt.Printf("Error opening input file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create job queue and wait group
	jobs := make(chan Job, maxQueueLength)
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(jobs, &wg, *outputDir)
	}

	// XML decoder
	decoder := xml.NewDecoder(file)
	var inPage bool
	var currentPage Page

	// Counters for statistics
	pageCount := 0
	articleCount := 0

	fmt.Println("Starting to parse Wikipedia XML dump...")

	// Parse XML
outer:
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Error reading XML token: %v\n", err)
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "page" {
				inPage = true
				currentPage = Page{}
			} else if inPage {
				if se.Name.Local == "title" {
					decoder.DecodeElement(&currentPage.Title, &se)
				} else if se.Name.Local == "ns" {
					decoder.DecodeElement(&currentPage.NS, &se)
				} else if se.Name.Local == "id" && currentPage.ID == 0 {
					decoder.DecodeElement(&currentPage.ID, &se)
				} else if se.Name.Local == "redirect" {
					var redirect Redirect
					for _, attr := range se.Attr {
						if attr.Name.Local == "title" {
							redirect.Title = attr.Value
							break
						}
					}
					currentPage.Redirect = &redirect
				} else if se.Name.Local == "text" {
					var text Text
					decoder.DecodeElement(&text, &se)
					currentPage.Revision.Text = text
				}
			}
		case xml.EndElement:
			if se.Name.Local == "page" {
				inPage = false
				pageCount++

				// Process only main namespace articles (ns = 0) that are not redirects
				if currentPage.NS == 0 && currentPage.Redirect == nil {
					articleCount++

					// Send job to worker pool
					jobs <- Job{
						Title: currentPage.Title,
						Text:  currentPage.Revision.Text.Content,
						ID:    currentPage.ID,
					}

					// Print progress every 10,000 articles
					if articleCount%10000 == 0 {
						fmt.Printf("Processed %d articles (total pages: %d)\n", articleCount, pageCount)
					}

					// Check if we've reached the limit
					if *limit > 0 && articleCount >= *limit {
						fmt.Printf("Reached limit of %d articles\n", *limit)
						break outer
					}
				}
			}
		}
	}

	// Close job channel and wait for all workers to finish
	close(jobs)
	wg.Wait()

	redirectCount := pageCount - articleCount
	fmt.Printf("Completed! Processed %d total pages, extracted %d articles (skipped %d redirects)\n", pageCount, articleCount, redirectCount)
}

// worker processes jobs from the queue
func worker(jobs <-chan Job, wg *sync.WaitGroup, outputDir string) {
	defer wg.Done()

	for job := range jobs {
		// Create sanitized filename
		filename := fmt.Sprintf("%x", sha256.Sum256([]byte(job.Title)))

		dirPath := filepath.Join(outputDir, filename[:3])

		// Create directory if it doesn't exist
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dirPath, err)
			continue
		}

		// Full path to output file
		filePath := filepath.Join(dirPath, filename+".md")

		// Create markdown file
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating file for '%s': %v\n", job.Title, err)
			continue
		}

		// Write markdown header with title
		_, err = file.WriteString(fmt.Sprintf("# %s\n\n", job.Title))
		if err != nil {
			fmt.Printf("Error writing header for '%s': %v\n", job.Title, err)
			if closeErr := file.Close(); closeErr != nil {
				fmt.Printf("Error closing file for '%s': %v\n", job.Title, closeErr)
			}
			continue
		}

		// Write content as-is
		_, err = file.WriteString(job.Text)
		if err != nil {
			fmt.Printf("Error writing content for '%s': %v\n", job.Title, err)
		}

		// Close the file and check for errors
		if err := file.Close(); err != nil {
			fmt.Printf("Error closing file for '%s': %v\n", job.Title, err)
		}
	}
}
