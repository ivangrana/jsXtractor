[# jsXtractor

Here's a professional and informative `README.md` for your Go tool:

---

# 🛠️ jsXtractor – JavaScript File Extractor & Wordlist Generator

`jsXtractor` is a lightweight, command-line tool written in Go that extracts JavaScript files from websites, downloads them, and optionally generates wordlists from their content. It’s ideal for security researchers, penetration testers, and bug bounty hunters performing reconnaissance or client-side analysis.

---

## 📋 Features

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
go install github.com/yourusername/jsextractor@latest
```

Or clone and build:
```bash
git clone https://github.com/yourusername/jsextractor.git
cd jsextractor
go build -o jsextractor main.go
```

---

## ▶️ Usage

```bash
./jsextractor [flags]
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

> You can also pipe domains via stdin.

---

## 🧪 Examples

### 1. Extract JS from a single domain
```bash
./jsXtractor -u example.com
```

### 2. Download JS files and save to directory
```bash
./jsXtractor -u https://example.com -dl -od js_files/
```

### 3. Generate wordlists from JS content
```bash
./jsXtractor -u example.com -dl -od js_files -w
```

### 4. Process list of domains from file
```bash
./jsXtractor -l domains.txt -dl -od js_files -w -o results.txt
```

### 5. Use with proxy
```bash
./jsXtractor -u example.com -p http://127.0.0.1:8080
```

### 6. Pipe domains via stdin
```bash
echo "example.com" | ./jsXtractor
```
or
```bash
cat domains.txt | ./jsXtractor -dl -od js_files
```

---

## 📂 Output

- **JS Files**: Saved in the specified directory (e.g., `js_files/`)
- **Wordlists**: Generated as `filename.js.wordlists`
- **Results**: Printed to stdout and optionally saved with `-o`

Example output:
```
Extracted domain: https://example.com
Results URL JS:
- https://example.com/static/app.js
   - File downloaded: js_files/app.js
   - Wordlists created: js_files/app.js.wordlists
```

---

## 🔐 Security Notes

- Automatically prepends `https://` if scheme is missing
- Resolves relative script URLs correctly
- Uses safe HTTP client with timeout and redirect limits
- Sanitizes filenames to avoid path traversal

---

## 🛠️ Future Enhancements (Planned)

- [ ] Add JSON output format
- [ ] Detect secrets (API keys, tokens) in JS
- [ ] Support concurrent domain processing
- [ ] Filter out CDN/common libraries (e.g., jQuery)
- [ ] Fingerprint JS frameworks/libraries

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
](https://bluelephant.com.br)
