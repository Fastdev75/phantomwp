package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

func main() {
	urlFlag := flag.String("l", "", "URL of the website to check")
	fileFlag := flag.String("f", "", "File containing URLs to check")
	outputFlag := flag.String("o", "", "Output file for WordPress detections")
	helpFlag := flag.Bool("help", false, "Show help information")

	flag.Parse()

	if *helpFlag {
		printHelp()
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

	if *urlFlag != "" {
		checkSingleURL(*urlFlag, output)
	} else if *fileFlag != "" {
		checkURLsFromFile(*fileFlag, output)
	} else {
		fmt.Println("Please provide either -l or -f flag. Use -help for more information.")
	}
}

func printHelp() {
	fmt.Println("PhantomWP - WordPress Detection Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  PhantomWP -l <url> [-o <output_file>]     Check a single URL")
	fmt.Println("  PhantomWP -f <file> [-o <output_file>]    Check URLs from a file")
	fmt.Println("  PhantomWP -help                           Show this help message")
	fmt.Println("\nOptions:")
	fmt.Println("  -l <url>          URL of the website to check")
	fmt.Println("  -f <file>         Path to a file containing URLs (one per line)")
	fmt.Println("  -o <output_file>  Output file for WordPress detections")
	fmt.Println("  -help             Show this help message")
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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			checkSingleURL(url, output)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}

func detectWordPress(url string) bool {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching URL: %v\n", err)
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
		resp, err := http.Get(url + path)
		if err == nil {
			defer resp.Body.Close()
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
