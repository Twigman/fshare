package httpapi

import (
	"net/http"
	"strings"
)

func (s *RESTService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	keyUUID, err := s.authorizeBearer(w, r)
	if err != nil {
		return
	}

	rUUID := strings.TrimPrefix(r.URL.Path, "/delete/")
	err = s.fileService.DeleteResourceByUUID(rUUID, keyUUID)
	if err != nil {
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
