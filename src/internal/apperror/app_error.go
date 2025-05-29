package apperror

import "net/http"

type FShareError struct {
	Code int
	Key  string
	Msg  string
}

func (e *FShareError) Error() string {
	return e.Msg
}

func (e *FShareError) Is(target error) bool {
	t, ok := target.(*FShareError)
	if !ok {
		return false
	}
	return e.Key == t.Key
}

func New(code int, key string, msg string) *FShareError {
	return &FShareError{
		Code: code,
		Key:  key,
		Msg:  msg,
	}
}

var (
	ErrFileInvalidFilename     = &FShareError{Code: http.StatusBadRequest, Key: "invalid_filename", Msg: "Filename not allowed"}
	ErrFileInvalidFilepath     = &FShareError{Code: http.StatusBadRequest, Key: "invalid_filepath", Msg: "Filepath not allowed"}
	ErrFileAlreadyExists       = &FShareError{Code: 11002, Key: "file_already_exists", Msg: "File already exists"}
	ErrFileAlreadyDeleted      = &FShareError{Code: 11003, Key: "file_already_deleted", Msg: "File already deleted"}
	ErrResourceNotFound        = &FShareError{Code: http.StatusNotFound, Key: "resource_not_found", Msg: "Resource not found"}
	ErrResourceResolvePath     = &FShareError{Code: http.StatusBadRequest, Key: "invalid_path", Msg: "Could not resolve path"}
	ErrCharsNotAllowed         = &FShareError{Code: 10001, Key: "invalid_characters", Msg: "Some characters are not allowed"}
	ErrDeleteHomeDirNotAllowed = &FShareError{Code: 11001, Key: "unauthorized_delete_home_dir", Msg: "Deleting the home directory is not allowed"}
	AuthorizationError         = &FShareError{Code: http.StatusUnauthorized, Key: "unauthorized", Msg: "Not authorized"}
)
