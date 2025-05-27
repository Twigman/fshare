package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/twigman/fshare/src/internal/apperror"
	"github.com/twigman/fshare/src/store"
)

func (s *RESTService) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyUUID, err := s.authorizeBearer(w, r)
	if err != nil {
		return
	}

	// upload limit
	if s.config.IsUploadLimited() {
		r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxFileSizeBytes())
		// space in RAM
		if err := r.ParseMultipartForm(s.config.MaxFileSizeBytes()); err != nil {
			writeJSONResponse(w, "File too large", http.StatusRequestEntityTooLarge)
			return
		}
	} else {
		// 32 MiB for RAM, rest will be created in /tmp
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			writeJSONResponse(w, "Upload error", http.StatusBadRequest)
			return
		}
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONResponse(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	// read fields
	isPrivate := r.FormValue("is_private") == "true"
	autoDelInH := r.FormValue("auto_del_in_h")

	i, err := strconv.Atoi(autoDelInH)
	if err != nil {
		i = 24
	}

	res := &store.Resource{
		Name:              header.Filename,
		IsPrivate:         isPrivate,
		APIKeyUUID:        keyUUID,
		AutoDeleteInHours: i,
	}

	file_uuid, err := s.resourceService.SaveUploadedFile(file, res)
	if err == apperror.ErrInvalidFilename {
		writeJSONResponse(w, apperror.ErrInvalidFilename.Msg, http.StatusBadRequest)
		return
	}
	if err != nil {
		writeJSONResponse(w, "Could not save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"uuid": file_uuid,
	})
}
