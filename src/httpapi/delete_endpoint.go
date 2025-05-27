package httpapi

import (
	"net/http"
	"strings"

	"github.com/twigman/fshare/src/internal/apperror"
)

func (s *RESTService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSONResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keyUUID, err := s.authorizeBearer(w, r)
	if err != nil {
		return
	}

	rUUID := strings.TrimPrefix(r.URL.Path, "/delete/")
	err = s.resourceService.DeleteResourceByUUID(rUUID, keyUUID)
	if err != nil {
		if err == apperror.ErrDeleteHomeDirNotAllowed {
			writeJSONResponse(w, "No permission to delete this object", http.StatusForbidden)
			return
		} else if err == apperror.ErrFileAlreadyDeleted {
			writeJSONResponse(w, "Resource not found", http.StatusNotFound)
			return
		} else if err == apperror.ErrResourceNotFound {
			writeJSONResponse(w, "Resource not found", http.StatusNotFound)
			return
		} else {
			writeJSONResponse(w, "Delete failed", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
