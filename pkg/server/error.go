package server

import "errors"

var (
	ErrInternalDataServiceError = errors.New("internal data service error")
	ErrInternalAPIServiceError  = errors.New("internal api service error")
	ErrParseParameter           = errors.New("parse parameter error")
	ErrInvalidParameter         = errors.New("invalid parameter")
)
