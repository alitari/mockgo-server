package utils

import "log"

type Logger struct {
	Verbose bool
}

func (l *Logger) LogAlways(formattedMessage string) {
	log.Print(formattedMessage)
}

func (l *Logger) LogWhenVerbose(formattedMessage string) {
	if l.Verbose {
		log.Print(formattedMessage)
	}
}
