package chat

import "errors"

var (
	errNonExistingChat = errors.New("reading from non existing chat")
	errBadMessage = errors.New("bad message")
)
