package httpapi

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *RESTService) RawResourceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONStatus(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// query paramters are not in the path
	file_uuid := strings.TrimPrefix(r.URL.Path, "/raw/")

	if !s.isValidSignedRequest(r, file_uuid) {
		writeJSONStatus(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	res, err := s.resourceService.GetResourceByUUID(file_uuid)
	if err != nil || res == nil || !res.IsFile || res.DeletedAt != nil || res.IsBroken {
		writeJSONStatus(w, http.StatusNotFound, "Not found")
		return
	}

	resPath, err := s.resourceService.BuildResourcePath(res)
	if err != nil {
		writeJSONStatus(w, http.StatusInternalServerError, "Could not resolve filepath")
		return
	}

	if _, err := os.Stat(resPath); err != nil {
		_ = s.resourceService.MarkResourceAsBroken(res.UUID)
		writeJSONStatus(w, http.StatusNotFound, "File not found")
		return
	}

	fileExt := filepath.Ext(res.Name)
	mimeType := mime.TypeByExtension(fileExt)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)

	if r.URL.Query().Get("download") == "true" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", res.Name))
	} else {
		w.Header().Set("X-Content-Type-Options", "nosniff")
	}

	http.ServeFile(w, r, resPath)
}
