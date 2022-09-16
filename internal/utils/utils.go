package utils

import (
	"log"
	"math/rand"
)

type Logger struct {
	Verbose                bool
	DebugRequestMatching   bool
	DebugResponseRendering bool
}

func (l *Logger) LogAlways(formattedMessage string) {
	log.Print(formattedMessage)
}

func (l *Logger) LogWhenVerbose(formattedMessage string) {
	if l.Verbose {
		log.Print(formattedMessage)
	}
}

func (l *Logger) LogWhenDebugRR(formattedMessage string) {
	if l.DebugResponseRendering {
		log.Print("(DEBUG) " + formattedMessage)
	}
}

func (l *Logger) LogWhenDebugRM(formattedMessage string) {
	if l.DebugRequestMatching {
		log.Print("(DEBUG) " + formattedMessage)
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letterBytes[rand.Int63() % int64(len(letterBytes))]
    }
    return string(b)
}

func Min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
