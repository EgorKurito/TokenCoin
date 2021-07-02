package util

import "log"

func LogErrHandle(err error) {
	log.Panicf("ERROR: %s\n", err)
}
