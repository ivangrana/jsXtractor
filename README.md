# jsXtractor

Here's a professional and informative `README.md` for your Go tool:

---

# 🛠️ JSExtractor – JavaScript File Extractor & Wordlist Generator

`JSExtractor` is a lightweight, command-line tool written in Go that extracts JavaScript files from websites, downloads them, and optionally generates wordlists from their content. It’s ideal for security researchers, penetration testers, and bug bounty hunters performing reconnaissance or client-side analysis.

---

## 📋 Features

- ✅ **Recursive Crawling** discover JS files on all reachable pages
- ✅ **Depth Control** configurable crawl depth
- ✅ **Scope Filtering** domain whitelisting/blacklisting via regex
- ✅ **Extract JS files** from `<script src="...">` tags
- 🔗 Supports domains with or without `http://`/`https://`
- 💾 **Download JS files** with retry logic
- 📂 Save files to a custom directory
- 📚 **Generate wordlists** (unique lowercase words) from JS content
- 🌐 Accept input from:
  - Command-line (`-u`)
  - File (`-l`)
  - Standard input (stdin)
- 🧩 Supports **HTTP/HTTPS proxies**
- 📝 Save results to an output file
- 🔄 Retry failed downloads
- 🧹 Clean, efficient, and concurrent-ready design

---

## 🚀 Installation

### Prerequisites
- [Go](https://golang.org/dl/) 1.18 or higher
- `git`

### Install
```bash
go build -o jsxtractor main.go
```

---

## ▶️ Usage

```bash
./jsxtractor [flags]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-u <url>` | Single domain URL | `""` |
| `-l <file>` | File containing list of domains (one per line) | `""` |
| `-o <file>` | Output file for results | `""` |
| `-dl` | Enable downloading of JS files | `true` |
| `-od <dir>` | Directory to save downloaded JS files | `""` |
| `-w` | Create wordlists from downloaded JS files | `false` |
| `-r <n>` | Number of download retries | `3` |
| `-p <proxy>` | Proxy URL (e.g. `http://127.0.0.1:8080`) | `""` |
| `-crawl` | Enable recursive crawling | `false` |
| `-depth <n>` | Crawl depth | `3` |
| `-scope <regex>` | Regex for in-scope domains | `""` |
| `-respect-robots`| Respect robots.txt (placeholder) | `false` |

---

## 🧪 Examples

### 1. Extract JS from a single domain
```bash
./jsxtractor -u example.com
```

### 2. Recursive crawl with depth 3
```bash
./jsxtractor -u https://example.com -crawl -depth 3
```

### 3. Crawl with scope regex
```bash
./jsxtractor -u https://example.com -crawl -scope ".*\.example\.com/api/.*"
```

### 4. Download JS files and save to directory
```bash
./jsxtractor -u https://example.com -crawl -dl -od js_files/
```

### 5. Generate wordlists from JS content
```bash
./jsxtractor -u example.com -dl -od js_files -w
```

### 6. Process list of domains from file
```bash
./jsxtractor -l domains.txt -dl -od js_files -w -o results.txt
```

### 7. Use with proxy
```bash
./jsxtractor -u example.com -p http://127.0.0.1:8080
```

---

## 📂 Output

- **JS Files**: Saved in the specified directory (e.g., `js_files/`)
- **Wordlists**: Generated as `filename.js.wordlists`
- **Results**: Printed to stdout and optionally saved with `-o`

Example output:
```
Starting crawl on: https://example.com (max depth: 3)

Page: https://example.com (Depth: 0)
- Found JS: https://example.com/static/app.js
   - File downloaded: js_files/app.js

Page: https://example.com/about (Depth: 1)
- Found JS: https://example.com/static/about.js
   - File downloaded: js_files/about.js
```

---

## 🔐 Security Notes

- Automatically prepends `https://` if scheme is missing
- Resolves relative script URLs correctly
- Uses safe HTTP client with timeout and redirect limits
- Sanitizes filenames to avoid path traversal

---

## 🛠️ Future Enhancements (Planned)

- [ ] **JS Rendering**: Headless browser integration for SPA (G-02)
- [ ] **Concurrency**: Worker pool with configurable parallelism (G-03)
- [ ] **Secret Detection**: Regex + entropy-based secret extraction (G-05)
- [ ] **Endpoint Extraction**: AST + regex endpoint harvesting (G-06)
- [ ] **Output Formats**: JSON, JSONL, CSV, Markdown, Silent (G-08)
- [ ] **Source Map Recovery**: `.js.map` detection + reconstruction (G-07)

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.

---

## Acknowledgments

- Uses [`goquery`](https://github.com/PuerkitoBio/goquery) for HTML parsing
- Inspired by tools like `subjs`, `gospider`, and `waybackurls`

---

## 📬 Feedback & Contributions

Issues, feature requests, and PRs are welcome!
Made with ❤️ for security research.
