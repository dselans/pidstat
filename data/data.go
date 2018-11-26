// +build dev
//
// Using 'dev' build tag will ensure that we read 'assets/*' from disk during
// development. At build time, binary will be built with '!dev' tag.
//

//go:generate go run -tags=dev ../assets_generate.go

package data

import (
	"net/http"
)

var Assets http.FileSystem = http.Dir("data/assets")
