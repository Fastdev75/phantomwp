package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

var (
	urlFlag     = flag.String("l", "", "URL of the website to check")
	fileFlag    = flag.String("f", "", "File containing URLs to check")
	outputFlag  = flag.String("o", "", "Output file for WordPress detections")
	timeoutFlag = flag.Int("t", 10, "Timeout for HTTP requests in seconds")
)

func main() {
	printASCIILogo()
	flag.Parse()

	if *urlFlag == "" && *fileFlag == "" {
		fmt.Println("Please provide either -l or -f flag. Use -help for more information.")
		return
	}

	var output io.Writer = os.Stdout
	if *outputFlag != "" {
		file, err := os.Create(*outputFlag)
		if err != nil {
			fmt.Printf("Error creating output file: %v\n", err)
			return
		}
		defer file.Close()
		output = file
	}

	start := time.Now()

	if *urlFlag != "" {
		checkSingleURL(*urlFlag, output)
	} else if *fileFlag != "" {
		checkURLsFromFile(*fileFlag, output)
	}

	duration := time.Since(start)
	fmt.Printf("\n%sTask completed in %v%s\n", colorGreen, duration, colorReset)
}

func printASCIILogo() {
	logo := `
    ____  __                 __              _       _______ 
   / __ \/ /_  ____ _____   / /_____  ____ _/ /  __ / / ___/ 
  / /_/ / __ \/ __ '/ __ \ / __/ __ \/ __ '/ _ \| / /\__ \  
 / ____/ / / / /_/ / / / // /_/ /_/ / /_/ / / / |/ /___/ /  
/_/   /_/ /_/\__,_/_/ /_/ \__/\____/\__, /_/_//___//____/   
                                   /____/                    
`
	fmt.Println(colorGreen + logo + colorReset)
}

func checkSingleURL(url string, output io.Writer) {
	if detectWordPress(url) {
		writeOutput(output, fmt.Sprintf("%sDetected WordPress%s: %s\n", colorGreen, colorReset, url))
	} else {
		writeOutput(output, fmt.Sprintf("%sNot WordPress%s: %s\n", colorYellow, colorReset, url))
	}
}

func checkURLsFromFile(filename string, output io.Writer) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	var wg sync.WaitGroup
	urlChan := make(chan string)

	// Start worker goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range urlChan {
				checkSingleURL(url, output)
			}
		}()
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urlChan <- url
		}
	}

	close(urlChan)
	wg.Wait()

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}

func detectWordPress(url string) bool {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	client := &http.Client{
		Timeout: time.Duration(*timeoutFlag) * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("%sError fetching URL%s: %v\n", colorRed, colorReset, err)
		return false
	}
	defer resp.Body.Close()

	// Check for WordPress-specific headers
	for key, values := range resp.Header {
		lowerKey := strings.ToLower(key)
		for _, value := range values {
			lowerValue := strings.ToLower(value)
			if (lowerKey == "x-powered-by" && strings.Contains(lowerValue, "wordpress")) ||
				(lowerKey == "link" && strings.Contains(lowerValue, "https://api.w.org/")) {
				return true
			}
		}
	}

	// Check for common WordPress paths
	commonPaths := []string{"/wp-content/", "/wp-includes/", "/wp-admin/"}
	for _, path := range commonPaths {
		resp, err := client.Get(url + path)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return true
			}
		}
	}

	return false
}

func writeOutput(w io.Writer, message string) {
	if f, ok := w.(*os.File); ok && f != os.Stdout {
		// Write to file without color codes
		fmt.Fprint(w, strings.ReplaceAll(strings.ReplaceAll(message, colorGreen, ""), colorYellow, ""))
	} else {
		// Write to stdout with color codes
		fmt.Fprint(w, message)
	}
}
