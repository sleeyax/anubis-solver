package url

import (
	_ "embed"
)

//go:embed assets/bundle.js
var polyfill string
