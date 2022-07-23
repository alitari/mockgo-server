package utils

import "log"

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
