package videocmprs

import "errors"

var (
	ErrNotRegistered      = errors.New("account not registered")
	ErrSomethingWentWrong = errors.New("something went wrong")
	ErrCantSendVideo      = errors.New("can't send video")
)
