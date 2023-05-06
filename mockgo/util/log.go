package util

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

/*
CreateLogger creates a logger with the given log level
*/
func CreateLogger(logLevel int) *zap.Logger {
	var logCfg zap.Config
	logType := os.Getenv("LOGTYPE")
	if logType == "prod" || logType == "production" {
		logCfg = zap.NewProductionConfig()
	} else {
		logCfg = zap.NewDevelopmentConfig()
	}
	logCfg.Level = zap.NewAtomicLevelAt(zapcore.Level(logLevel))
	logger, err := logCfg.Build()
	if err != nil {
		log.Fatal(err)
	}
	return logger
}
