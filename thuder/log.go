package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/xiegeo/thuder"
	"gopkg.in/natefinch/lumberjack.v2"
)

//logger write to both log file and stdout
func logger(hc *thuder.HostConfig) io.Writer {
	logger := &lumberjack.Logger{
		Filename:   filepath.Join(hc.UniqueDirectory(), "log"),
		MaxSize:    1,
		MaxBackups: 2,
	}
	return io.MultiWriter(os.Stdout, logger)
}
