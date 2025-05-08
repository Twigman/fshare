package httpapi

import (
	"path/filepath"
	"strings"
)

var imageTypeWhitelist = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".avif": true,
}

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

func isRenderableTextFile(ext string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	_, ok := highlightExtWhitelist[ext]
	return ok
}

func isRenderableImageFile(ext string) bool {
	ext = strings.ToLower(filepath.Ext(ext))
	return imageTypeWhitelist[ext]
}

func getLangClass(ext string) string {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	lang, ok := highlightExtWhitelist[ext]
	if !ok {
		return "plaintext"
	}
	return lang
}
