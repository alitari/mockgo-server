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
func CreateLogger(logLevelTxt string) *zap.Logger {
	var logCfg zap.Config
	logType := os.Getenv("LOGTYPE")
	if logType == "prod" || logType == "production" {
		logCfg = zap.NewProductionConfig()
	} else {
		logCfg = zap.NewDevelopmentConfig()
	}
	level, err := zapcore.ParseLevel(logLevelTxt)
	if err != nil {
		log.Fatal(err)
	}
	logCfg.Level = zap.NewAtomicLevelAt(level)
	logger, err := logCfg.Build()
	if err != nil {
		log.Fatal(err)
	}
	return logger
}
