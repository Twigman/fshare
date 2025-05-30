package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/twigman/fshare/src/internal/apperror"
	"github.com/twigman/fshare/src/store"
)

func (s *RESTService) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONStatus(w, http.StatusMethodNotAllowed, "Method not allowed")
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
			writeJSONStatus(w, http.StatusRequestEntityTooLarge, "File too large")
			return
		}
	} else {
		// 32 MiB for RAM, rest will be created in /tmp
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			writeJSONStatus(w, http.StatusBadRequest, "Upload error")
			return
		}
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONStatus(w, http.StatusBadRequest, "Invalid file")
		return
	}
	defer file.Close()
	// read fields
	isPrivate := r.FormValue("is_private") == "true"

	// handle TTL
	autoDelInRaw := r.FormValue("auto_del_in")
	var autoDeleteTime time.Time
	var autoDeleteAt *time.Time

	if autoDelInRaw == "" {
		autoDeleteAt = nil // default no auto delete
	} else if strings.HasSuffix(autoDelInRaw, "d") {
		daysStr := strings.TrimSuffix(autoDelInRaw, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil || days < 0 {
			autoDeleteTime = time.Now().Add(24 * time.Hour).UTC() // fallback
			autoDeleteAt = &autoDeleteTime
		} else {
			autoDeleteTime = time.Now().Add(time.Duration(days) * 24 * time.Hour).UTC()
			autoDeleteAt = &autoDeleteTime
		}
	} else {
		autoDeleteDur, err := time.ParseDuration(autoDelInRaw)
		if err != nil || autoDeleteDur < 0 {
			autoDeleteDur = 24 * time.Hour // fallback
		}
		autoDeleteTime = time.Now().Add(autoDeleteDur).UTC()
		autoDeleteAt = &autoDeleteTime
	}

	res := &store.Resource{
		Name:         header.Filename,
		IsPrivate:    isPrivate,
		APIKeyUUID:   keyUUID,
		AutoDeleteAt: autoDeleteAt,
	}

	file_uuid, err := s.resourceService.SaveUploadedFile(file, res, true)
	if err == apperror.ErrFileInvalidFilename {
		writeJSONStatus(w, http.StatusBadRequest, apperror.ErrFileInvalidFilename.Msg)
		return
	}
	if err != nil {
		writeJSONStatus(w, http.StatusInternalServerError, "Could not save file")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"uuid": file_uuid,
	})
}
