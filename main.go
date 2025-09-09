package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func downloadFile(urlStr, outputDir string, retries int) string {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var resp *http.Response
	var err error

	for i := 0; i < retries; i++ {
		resp, err = client.Get(urlStr)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if i < retries-1 {
			log.Printf("Error downloading %s (attempt %d): %v", urlStr, i+1, err)
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Download failed after %d retries: %s", retries, urlStr)
		return ""
	}
	defer resp.Body.Close()

	// Extract filename from URL
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return ""
	}
	pathParts := strings.Split(u.Path, "/")
	filename := pathParts[len(pathParts)-1]
	if filename == "" {
		filename = "index.js"
	}

	// Create output directory if needed
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating directory: %v", err)
		return ""
	}

	filePath := filepath.Join(outputDir, filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return ""
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, resp.Body); err != nil {
		log.Printf("Error writing file: %v", err)
		return ""
	}

	return filePath
}

func createWordlists(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return ""
	}

	// Extract words using regex
	re := regexp.MustCompile(`\b\w+\b`)
	matches := re.FindAllString(string(content), -1)

	wordSet := make(map[string]struct{})
	for _, word := range matches {
		wordSet[strings.ToLower(word)] = struct{}{}
	}

	// Write unique words to file
	wordlistPath := filePath + ".wordlists"
	wordlistFile, err := os.Create(wordlistPath)
	if err != nil {
		log.Printf("Error creating wordlist: %v", err)
		return ""
	}
	defer wordlistFile.Close()

	for word := range wordSet {
		if _, err := wordlistFile.WriteString(word + "\n"); err != nil {
			log.Printf("Error writing wordlist: %v", err)
			return ""
		}
	}

	return wordlistPath
}

func extractJSFromDomain(domain string, downloadFiles bool, outputDir string, createLists bool, retries int) []string {
	results := []string{}
	results = append(results, fmt.Sprintf("Extracted domain: %s", domain))

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(domain)
	if err != nil {
		results = append(results, fmt.Sprintf("Error fetching domain: %v", err))
		return results
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		results = append(results, fmt.Sprintf("HTTP error: %d", resp.StatusCode))
		return results
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		results = append(results, fmt.Sprintf("Error parsing HTML: %v", err))
		return results
	}

	uniqueURLs := make(map[string]struct{})
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			absoluteURL, err := url.Parse(src)
			if err != nil {
				return
			}
			base, err := url.Parse(domain)
			if err != nil {
				return
			}
			absoluteURL = base.ResolveReference(absoluteURL)
			uniqueURLs[absoluteURL.String()] = struct{}{}
		}
	})

	if len(uniqueURLs) == 0 {
		results = append(results, "No JavaScript files found")
		return results
	}

	results = append(results, "Results URL JS:")
	for jsURL := range uniqueURLs {
		results = append(results, fmt.Sprintf("- %s", jsURL))

		if downloadFiles && outputDir != "" {
			filePath := downloadFile(jsURL, outputDir, retries)
			if filePath != "" {
				results = append(results, fmt.Sprintf("   - File downloaded: %s", filePath))

				if createLists {
					wordlistPath := createWordlists(filePath)
					if wordlistPath != "" {
						results = append(results, fmt.Sprintf("   - Wordlists created: %s", wordlistPath))
					} else {
						results = append(results, "   - Wordlist creation failed")
					}
				}
			} else {
				results = append(results, "   - Download failed")
			}
		}
	}

	return results
}

func main() {
	var (
		urlStr         = flag.String("u", "", "Single domain URL")
		listFile       = flag.String("l", "", "File containing list of domains")
		outputFile     = flag.String("o", "", "Output file for results")
		download       = flag.Bool("dl", true, "Enable file download")
		retries        = flag.Int("r", 3, "Number of download retries")
		outputDir      = flag.String("od", "", "Directory for downloaded files")
		createWordlist = flag.Bool("w", false, "Create wordlists from files")
		proxy          = flag.String("p", "", "Proxy server URL")
	)
	flag.Parse()

	// Set HTTP proxy if specified
	if *proxy != "" {
		proxyURL, err := url.Parse(*proxy)
		if err != nil {
			log.Fatalf("Invalid proxy URL: %v", err)
		}
		http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	// Process input domains
	domains := []string{}

	// Read from stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			domains = append(domains, scanner.Text())
		}
	}

	// Process command-line arguments
	if *urlStr != "" {
		domains = append(domains, *urlStr)
	}

	if *listFile != "" {
		file, err := os.Open(*listFile)
		if err != nil {
			log.Fatalf("Error opening list file: %v", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			domains = append(domains, scanner.Text())
		}
	}

	// Create output directory if specified
	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}
	}

	// Process all domains
	var outputBuffer bytes.Buffer
	for _, domain := range domains {
		results := extractJSFromDomain(
			domain,
			*download,
			*outputDir,
			*createWordlist,
			*retries,
		)

		for _, line := range results {
			fmt.Println(line)
			outputBuffer.WriteString(line + "\n")
		}
		fmt.Println()
		outputBuffer.WriteString("\n")
	}

	// Write to output file if specified
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, outputBuffer.Bytes(), 0644); err != nil {
			log.Fatalf("Error writing output file: %v", err)
		}
	}
}
