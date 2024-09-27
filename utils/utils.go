package utils

import (
	"encoding/csv"
	"os"
)

func IsPathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func CreateCsv(path string) error {
	if IsPathExists(path) {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	writer := csv.NewWriter(file)
	err = writer.Write([]string{"uid", "createDate", "startTime", "endTime", "cpu_num", "mem", "gpu_num", "worker_num"})
	writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
