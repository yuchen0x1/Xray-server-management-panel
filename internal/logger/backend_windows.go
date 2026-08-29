//go:build windows || plan9

package logger

import (
	"os"

	"github.com/op/go-logging"
)

func initDefaultBackend() logging.Backend {
	return logging.NewBackendFormatter(
		logging.NewLogBackend(os.Stderr, "", 0),
		newFormatter(true),
	)
}
