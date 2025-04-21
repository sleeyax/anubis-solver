package worker

import (
	_ "embed"
)

//go:embed assets/worker.js
var polyfill string
