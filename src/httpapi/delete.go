package httpapi

import (
	"net/http"
	"strings"
)

func (s *RESTService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyUUID, err := s.authorizeBearer(w, r)
	if err != nil {
		return
	}

	rUUID := strings.TrimPrefix(r.URL.Path, "/delete/")
	err = s.resourceService.DeleteResourceByUUID(rUUID, keyUUID)
	if err != nil {
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
