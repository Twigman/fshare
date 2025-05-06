package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/twigman/fshare/src/store"
)

func (s *RESTService) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		http.Error(w, "Invalid Authorization scheme", http.StatusUnauthorized)
		return
	}

	// validate API key
	apiKey := strings.TrimPrefix(authHeader, prefix)
	key_uuid, err := s.apiKeyService.GetUUIDForAPIKey(apiKey)
	if err != nil {
		http.Error(w, "Authorization failed", http.StatusUnauthorized)
		return
	}

	// upload limit
	if s.config.IsUploadLimited() {
		r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxFileSizeBytes())
		// space in RAM
		if err := r.ParseMultipartForm(s.config.MaxFileSizeBytes()); err != nil {
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
	// read fields
	isPrivate := r.FormValue("is_private") == "true"
	//folder := r.FormValue("folder") == ""
	autoDelInH := r.FormValue("auto_del_in_h")

	i, err := strconv.Atoi(autoDelInH)
	if err != nil {
		http.Error(w, "Invalid value for auto_del_in_h", http.StatusBadRequest)
		return
	}

	res := &store.Resource{
		Name:              header.Filename,
		IsPrivate:         isPrivate,
		APIKeyUUID:        key_uuid,
		AutoDeleteInHours: i,
	}

	file_uuid, err := s.fileService.SaveUploadedFile(file, res)
	if err != nil {
		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"uuid": file_uuid,
	})
}
