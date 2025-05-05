package httpapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/twigman/fshare/src/store"
)

func (s *RESTService) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
	apiKey := r.FormValue("api_key")
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
		OwnerHashedKey:    store.HashAPIKey(apiKey),
		AutoDeleteInHours: i,
	}

	file_uuid, err := s.fileService.SaveUploadedFile(file, res)
	if err != nil {
		http.Error(w, "Could not save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Uploaded successfully: %s\n", file_uuid)
}
