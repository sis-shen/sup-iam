package repository

import "errors"

var (
	ErrNotFound       = errors.New("repository: not found")
	ErrAlreadyExists  = errors.New("repository: already exists")
	ErrConflict       = errors.New("repository: conflict")
	ErrInvalidInput   = errors.New("repository: invalid input")
	ErrStorageFailure = errors.New("repository: storage failure")
)
