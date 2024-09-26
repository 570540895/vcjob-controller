package utils

import (
	"os"
	"time"
)

func IsPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func UTCTransLocal(utcTime string) string {
	t, _ := time.Parse("2006-01-02T15:04:05Z", utcTime)
	loc, _ := time.LoadLocation("Asia/Shanghai")
	bjTime := t.In(loc)
	return bjTime.Format("2006-01-02 15:04:05")
}
