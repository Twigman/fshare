package httpapi

import (
	"path/filepath"
	"strings"
)

var imageTypeWhitelistDefault = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
	".avif": true,
}

var imageTypeWhitelistTrusted = func() map[string]bool {
	trusted := make(map[string]bool)
	for k, v := range imageTypeWhitelistDefault {
		trusted[k] = v
	}
	// add types
	trusted[".svg"] = true
	trusted[".ico"] = true
	trusted[".tif"] = true
	trusted[".tiff"] = true
	trusted[".heic"] = true
	return trusted
}()

var highlightExtWhitelistDefault = map[string]string{
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

var highlightExtWhitelistTrusted = func() map[string]string {
	trusted := make(map[string]string)
	// base
	for k, v := range highlightExtWhitelistDefault {
		trusted[k] = v
	}
	// add extentions
	trusted["rs"] = "rust"
	trusted["swift"] = "swift"
	trusted["conf"] = "ini"
	trusted["bat"] = "dos"
	trusted["ps1"] = "powershell"
	trusted["tsx"] = "typescript"
	trusted["jsx"] = "javascript"
	trusted["vue"] = "xml"
	trusted["asm"] = "x86asm"
	trusted["log"] = "plaintext"
	return trusted
}()

func isRenderableTextFile(ext string, trusted bool) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	if trusted {
		_, ok := highlightExtWhitelistTrusted[ext]
		return ok
	}
	_, ok := highlightExtWhitelistDefault[ext]
	return ok
}

func isRenderableImageFile(ext string, trusted bool) bool {
	ext = strings.ToLower(filepath.Ext(ext))
	if trusted {
		return imageTypeWhitelistTrusted[ext]
	}
	return imageTypeWhitelistDefault[ext]
}

func getLangClass(ext string, trusted bool) string {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	if trusted {
		if lang, ok := highlightExtWhitelistTrusted[ext]; ok {
			return lang
		}
	} else {
		if lang, ok := highlightExtWhitelistDefault[ext]; ok {
			return lang
		}
	}
	return "plaintext"
}

func isBrowserRenderableFile(ext string) bool {
	ext = strings.ToLower(filepath.Ext(ext))
	switch ext {
	//case ".pdf", ".svg", ".mp4", ".webm", ".mp3", ".ogg", ".wav":
	case ".pdf", ".svg":
		return true
	default:
		return false
	}
}
