package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func (h *HTTPHandler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// upload limit
	if h.config.IsUploadLimited() {
		r.Body = http.MaxBytesReader(w, r.Body, h.config.MaxFileSizeBytes())
		// space in RAM
		if err := r.ParseMultipartForm(h.config.MaxFileSizeBytes()); err != nil {
			http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
			return
		}
	} else {
		// 32 MiB for RAM, rest will be created in /tmp
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Upload error", http.StatusBadRequest)
			return
		}
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// remove paths (../../file.txt -> file.txt)
	filename := filepath.Base(header.Filename)
	dst, err := os.Create(filepath.Join(h.config.UploadPath, filename))
	if err != nil {
		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Could not write file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Uploaded successfully: %s\n", header.Filename)
}
