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
	ErrInvalidFilename = &FShareError{Code: http.StatusBadRequest, Key: "invalid_filename", Msg: "Filename not allowed"}
	ErrInvalidFilepath = &FShareError{Code: http.StatusBadRequest, Key: "invalid_filepath", Msg: "Filepath not allowed"}
	ErrResolvePath     = &FShareError{Code: http.StatusBadRequest, Key: "invalid_filepath", Msg: "Could not resolve filepath"}
	ErrCharsNotAllowed = &FShareError{Code: 10001, Key: "invalid_characters", Msg: "Some characters are not allowed"}
)
