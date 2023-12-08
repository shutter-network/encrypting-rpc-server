package server

import (
	"os"

	"github.com/rs/zerolog"
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
