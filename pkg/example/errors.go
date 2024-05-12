package example

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrNotConnected = errors.New("not connected")
var ErrUnknownRepositoryType = errors.New("unknown repository type")

func IsNotFound(err error) bool {
	return err == ErrNotFound
}

func IsAlreadyExists(err error) bool {
	return err == ErrAlreadyExists
}

func IsNotConnected(err error) bool {
	return err == ErrNotConnected
}

func IsUnknownRepositoryType(err error) bool {
	return err == ErrUnknownRepositoryType
}
