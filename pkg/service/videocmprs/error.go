package videocmprs

import "errors"

var (
	// ErrNotRegistered uses when account is not registered
	ErrNotRegistered = errors.New("account not registered")
	// ErrSomethingWentWrong uses when unexpected error appear
	ErrSomethingWentWrong = errors.New("something went wrong")
	// ErrCantSendVideo uses when can't send video
	ErrCantSendVideo = errors.New("can't send video")
)
