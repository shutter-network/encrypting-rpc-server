package rpc

import (
	"github.com/rs/zerolog"
	"os"
)

func setUpLogger() zerolog.Logger {
	logger := zerolog.New(os.Stdout).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Logger()
	return logger
}

var Logger = setUpLogger()
