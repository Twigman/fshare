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

func (s *RESTService) ResourceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file_uuid := strings.TrimPrefix(r.URL.Path, "/r/")

	res, err := s.resourceService.GetResourceByUUID(file_uuid)
	if err != nil || res == nil || !res.IsFile || res.DeletedAt != nil || res.IsBroken {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if res.IsPrivate {
		keyUUID, err := s.authorizeBearer(w, r)
		if err != nil {
			return
		}

		if res.APIKeyUUID != keyUUID {
			http.Error(w, "Authorization failed", http.StatusUnauthorized)
			return
		}
	}

	resPath := filepath.Join(s.config.UploadPath, res.APIKeyUUID, res.Name)
	fileExt := filepath.Ext(res.Name)

	absPath, err := filepath.Abs(resPath)
	basePath, err2 := filepath.Abs(s.config.UploadPath)

	if err != nil || err2 != nil || !strings.HasPrefix(absPath, basePath) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// detect mime type
	mimeType := mime.TypeByExtension(fileExt)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// present source code in HTML with highlighting
	content, err := os.ReadFile(resPath)
	if err != nil {
		_ = s.resourceService.MarkResourceAsBroken(res.UUID)

		http.Error(w, "Could not read file", http.StatusInternalServerError)
		return
	}

	if isRenderableTextFile(fileExt) {
		nonce := generateNonce()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", fmt.Sprintf(
			"default-src 'none'; script-src https://cdnjs.cloudflare.com 'nonce-%s'; style-src https://cdnjs.cloudflare.com 'nonce-%s'; img-src 'self'; object-src 'none'; base-uri 'none';",
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
	} else if isRenderableImageFile(fileExt) {
		// present images in browser
		if strings.HasPrefix(mimeType, "image/") {
			w.Header().Set("Content-Type", mimeType)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeFile(w, r, resPath)
			return
		}
	} else {
		// force download
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", res.Name))
		http.ServeFile(w, r, resPath)
		return
	}
}

func generateNonce() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "staticfallback"
	}
	return base64.StdEncoding.EncodeToString(b)
}
