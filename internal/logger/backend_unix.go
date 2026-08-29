//go:build !windows && !plan9

package logger

import (
	"fmt"
	"os"

	"github.com/op/go-logging"
)

func initDefaultBackend() logging.Backend {
	if backend, err := logging.NewSyslogBackend(""); err == nil {
		return logging.NewBackendFormatter(backend, newFormatter(false))
	} else {
		fmt.Fprintf(os.Stderr, "syslog backend disabled: %v\n", err)
		return logging.NewBackendFormatter(
			logging.NewLogBackend(os.Stderr, "", 0),
			newFormatter(os.Getppid() > 0),
		)
	}
}
