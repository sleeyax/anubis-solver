package blob

import (
	_ "embed"
)

//go:embed assets/blob.js
var polyfill string
