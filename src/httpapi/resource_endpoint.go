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
	"time"
)

func (s *RESTService) ResourceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file_uuid := strings.TrimPrefix(r.URL.Path, "/r/")

	res, err := s.resourceService.GetResourceByUUID(file_uuid)
	if err != nil || res == nil || !res.IsFile || res.DeletedAt != nil || res.IsBroken {
		writeJSONResponse(w, "Not found", http.StatusNotFound)
		return
	}

	if res.IsPrivate {
		keyUUID, err := s.authorizeBearer(w, r)
		if err != nil {
			return
		}

		if res.APIKeyUUID != keyUUID {
			writeJSONResponse(w, "Authorization failed", http.StatusUnauthorized)
			return
		}
	}

	resPath, err := s.resourceService.BuildResourcePath(res)
	if err != nil {
		writeJSONResponse(w, "Could not resolve filepath", http.StatusInternalServerError)
		return
	}
	fileExt := filepath.Ext(res.Name)

	// detect mime type
	mimeType := mime.TypeByExtension(fileExt)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// present source code in HTML with highlighting
	content, err := os.ReadFile(resPath)
	if err != nil {
		_ = s.resourceService.MarkResourceAsBroken(res.UUID)

		writeJSONResponse(w, "Could not read file", http.StatusInternalServerError)
		return
	}

	keyIsHighlyTrusted, err := s.apiKeyService.IsAPIKeyHighlyTrusted(res.APIKeyUUID)
	if err != nil {
		// proceed
		keyIsHighlyTrusted = false
	}

	if isRenderableTextFile(fileExt, keyIsHighlyTrusted) {
		renderText(w, getLangClass(fileExt, keyIsHighlyTrusted), string(content))
	} else if isRenderableImageFile(fileExt, keyIsHighlyTrusted) {
		// present images in browser
		if strings.HasPrefix(mimeType, "image/") {
			w.Header().Set("Content-Type", mimeType)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			http.ServeFile(w, r, resPath)
			return
		}
	} else if keyIsHighlyTrusted && isBrowserRenderableFile(fileExt) {
		s.renderMediaViewer(w, res.UUID, mimeType)
		return
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

func (s *RESTService) renderMediaViewer(w http.ResponseWriter, rUUID string, mimeType string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	expiry := time.Now().Add(30 * time.Second)
	signedURL, err := s.generateSignedURL("/raw", rUUID, expiry)
	if err != nil {
		writeJSONResponse(w, "Failed to generate signed URL", http.StatusInternalServerError)
		return
	}

	escapedPath := html.EscapeString(signedURL)

	var embed string
	switch {
	case strings.HasPrefix(mimeType, "application/pdf"):
		embed = fmt.Sprintf(`<iframe src="%s" width="100%%" height="100%%" style="border:none;"></iframe>`, escapedPath)
	case strings.HasPrefix(mimeType, "image/svg"):
		embed = fmt.Sprintf(`<img src="%s" style="max-width:100%%; max-height:100%%;">`, escapedPath)
	default:
		http.Error(w, "Unsupported viewer", http.StatusUnsupportedMediaType)
		return
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
	<html><head><meta charset="utf-8"><title>Viewer</title></head>
	<body style="margin:0; background:#111; color:#fff; display:flex; align-items:center; justify-content:center; height:100vh;">
	%s
	</body></html>`, embed)
}

func renderText(w http.ResponseWriter, langClass string, content string) {
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
		`, nonce, nonce, langClass, html.EscapeString(content))
}
