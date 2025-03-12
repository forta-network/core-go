package utils

import "github.com/sirupsen/logrus"

// FatalIfError logs a fatal error and exits the process if the error is not nil.
func FatalIfError(err error) {
	if err != nil {
		logrus.WithError(err).Fatal("fatal error - exiting process")
	}
}
