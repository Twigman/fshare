package httpapi

import (
	"net/http"
	"strings"

	"github.com/twigman/fshare/src/internal/apperror"
)

func (s *RESTService) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSONStatus(w, http.StatusMethodNotAllowed, "Method not allowed")
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
			writeJSONStatus(w, http.StatusForbidden, "No permission to delete this object")
			return
		} else if err == apperror.ErrFileAlreadyDeleted {
			writeJSONStatus(w, http.StatusNotFound, "Resource not found")
			return
		} else if err == apperror.ErrResourceNotFound {
			writeJSONStatus(w, http.StatusNotFound, "Resource not found")
			return
		} else if err == apperror.AuthorizationError {
			// should not occur, as it is validated beforehand
			writeJSONStatus(w, apperror.AuthorizationError.Code, "Unauthorized")
			return
		} else {
			writeJSONStatus(w, http.StatusInternalServerError, "Delete failed")
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
