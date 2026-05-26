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

type crawlItem struct {
	url   string
	depth int
}

func isInScope(targetURL *url.URL, baseHost string, scopeRegex *regexp.Regexp) bool {
	if scopeRegex != nil {
		return scopeRegex.MatchString(targetURL.String())
	}
	// Default: same host
	return targetURL.Host == baseHost
}

func crawlDomain(startURL string, maxDepth int, scopeRegex *regexp.Regexp, downloadFiles bool, outputDir string, createLists bool, retries int) []string {
	results := []string{}
	results = append(results, fmt.Sprintf("Starting crawl on: %s (max depth: %d)", startURL, maxDepth))

	u, err := url.Parse(startURL)
	if err != nil {
		results = append(results, fmt.Sprintf("Error parsing start URL: %v", err))
		return results
	}
	baseHost := u.Host

	visited := make(map[string]struct{})
	discoveredJS := make(map[string]struct{})
	queue := []crawlItem{{url: startURL, depth: 0}}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if _, exists := visited[item.url]; exists {
			continue
		}
		visited[item.url] = struct{}{}

		if item.depth > maxDepth {
			continue
		}

		resp, err := client.Get(item.url)
		if err != nil {
			log.Printf("Error fetching %s: %v", item.url, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}

		currentPageURL, err := url.Parse(item.url)
		if err != nil {
			log.Printf("Error parsing current page URL %s: %v", item.url, err)
			continue
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Error parsing HTML from %s: %v", item.url, err)
			continue
		}

		results = append(results, fmt.Sprintf("\nPage: %s (Depth: %d)", item.url, item.depth))

		// Extract JS
		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				jsURL, err := url.Parse(src)
				if err != nil {
					return
				}
				jsURL = currentPageURL.ResolveReference(jsURL)
				jsURLStr := jsURL.String()

				if _, exists := discoveredJS[jsURLStr]; !exists {
					discoveredJS[jsURLStr] = struct{}{}
					results = append(results, fmt.Sprintf("- Found JS: %s", jsURLStr))

					if downloadFiles && outputDir != "" {
						filePath := downloadFile(jsURLStr, outputDir, retries)
						if filePath != "" {
							results = append(results, fmt.Sprintf("   - File downloaded: %s", filePath))
							if createLists {
								wordlistPath := createWordlists(filePath)
								if wordlistPath != "" {
									results = append(results, fmt.Sprintf("   - Wordlists created: %s", wordlistPath))
								}
							}
						}
					}
				}
			}
		})

		// Extract Links for next depth
		if item.depth < maxDepth {
			doc.Find("a").Each(func(i int, s *goquery.Selection) {
				if href, exists := s.Attr("href"); exists {
					linkURL, err := url.Parse(href)
					if err != nil {
						return
					}
					linkURL = currentPageURL.ResolveReference(linkURL)

					// Remove fragments
					linkURL.Fragment = ""

					if isInScope(linkURL, baseHost, scopeRegex) {
						linkStr := linkURL.String()
						if _, exists := visited[linkStr]; !exists {
							queue = append(queue, crawlItem{url: linkStr, depth: item.depth + 1})
						}
					}
				}
			})
		}
	}

	return results
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
		crawl          = flag.Bool("crawl", false, "Enable recursive crawling")
		depth          = flag.Int("depth", 3, "Crawl depth")
		scope          = flag.String("scope", "", "Regex for in-scope domains")
		respectRobots  = flag.Bool("respect-robots", false, "Respect robots.txt (not implemented yet)")
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

	var scopeRegex *regexp.Regexp
	if *scope != "" {
		var err error
		scopeRegex, err = regexp.Compile(*scope)
		if err != nil {
			log.Fatalf("Invalid scope regex: %v", err)
		}
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
		var results []string
		if *crawl {
			results = crawlDomain(
				domain,
				*depth,
				scopeRegex,
				*download,
				*outputDir,
				*createWordlist,
				*retries,
			)
		} else {
			results = extractJSFromDomain(
				domain,
				*download,
				*outputDir,
				*createWordlist,
				*retries,
			)
		}

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
