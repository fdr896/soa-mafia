package filesystem

import "errors"

var ErrUnsupportedContentType = errors.New("unsupported content type: expected 'image/jpeg' or 'image/png'")
