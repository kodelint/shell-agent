package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(debug, verbose bool) {
	log = logrus.New()

	// Set log level
	if debug {
		log.SetLevel(logrus.DebugLevel)
	} else if verbose {
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(logrus.WarnLevel)
	}

	// Set formatter
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	log.SetOutput(os.Stdout)
}

func GetLogger() *logrus.Logger {
	if log == nil {
		InitLogger(false, false)
	}
	return log
}
