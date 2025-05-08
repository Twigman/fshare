package httpapi

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *RESTService) ViewHandler(w http.ResponseWriter, r *http.Request) {
	file_uuid := strings.TrimPrefix(r.URL.Path, "/v/")

	res, err := s.resourceService.GetResourceByUUID(file_uuid)
	if err != nil || res == nil || !res.IsFile || res.DeletedAt != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if res.IsPrivate {
		// TODO check api key

		http.Error(w, "Authorization failed", http.StatusUnauthorized)
		return
	}

	resPath := filepath.Join(s.config.UploadPath, res.APIKeyUUID, res.Name)
	fileExt := filepath.Ext(res.Name)

	// detect mime type
	mimeType := mime.TypeByExtension(fileExt)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// present images in browser
	if strings.HasPrefix(mimeType, "image/") {
		w.Header().Set("Content-Type", mimeType)
		http.ServeFile(w, r, resPath)
		return
	}

	// present source code in HTML with highlighting
	content, err := os.ReadFile(resPath)
	if err != nil {
		http.Error(w, "Could not read file", http.StatusInternalServerError)
		return
	}

	nonce := generateNonce()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", fmt.Sprintf(
		"default-src 'self'; script-src https://cdnjs.cloudflare.com 'nonce-%s'; style-src https://cdnjs.cloudflare.com 'nonce-%s';",
		nonce, nonce,
	))

	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html lang="en">
		<head>
		<meta charset="utf-8">
		<title>Viewer</title>
		<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css">
		<style nonce="%s">
			body {
			margin: 0;
			background-color: #0d1117;
			color: #c9d1d9;
			font-family: monospace;
			}

			pre {
			margin: 0;
			padding: 1em;
			background: #0d1117;
			box-shadow: none;
			border: none;
			overflow-x: auto;
			}

			code {
			background: none;
			color: inherit;
			}
		</style>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js" defer></script>
		<script nonce="%s">
			window.addEventListener('DOMContentLoaded', function () {
			document.querySelectorAll('pre code').forEach(function (el) {
				hljs.highlightElement(el);
			});
			});
		</script>
		</head>
		<body>
		<pre><code class="language-%s">%s</code></pre>
		</body>
		</html>
		`, nonce, nonce, getLangClass(fileExt), html.EscapeString(string(content)))
}

func getLangClass(ext string) string {
	var highlightExtWhitelist = map[string]string{
		"go":         "go",
		"js":         "javascript",
		"ts":         "typescript",
		"json":       "json",
		"html":       "html",
		"css":        "css",
		"scss":       "scss",
		"py":         "python",
		"sh":         "bash",
		"bash":       "bash",
		"c":          "c",
		"cpp":        "cpp",
		"h":          "cpp",
		"cs":         "csharp",
		"java":       "java",
		"kt":         "kotlin",
		"rb":         "ruby",
		"php":        "php",
		"sql":        "sql",
		"xml":        "xml",
		"yaml":       "yaml",
		"yml":        "yaml",
		"md":         "markdown",
		"toml":       "toml",
		"ini":        "ini",
		"dockerfile": "dockerfile",
		"makefile":   "makefile",
		"txt":        "plaintext",
	}

	ext = strings.TrimPrefix(ext, ".")
	langClass, ok := highlightExtWhitelist[strings.ToLower(ext)]

	if !ok {
		langClass = "plaintext"
	}

	return langClass
}

func generateNonce() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "staticfallback"
	}
	return base64.StdEncoding.EncodeToString(b)
}
