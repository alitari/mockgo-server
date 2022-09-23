package logging

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"runtime"
)

type LogLevel int64

const (
	OnlyError LogLevel = iota
	Verbose
	Debug
)

func ParseLogLevel(ll int) LogLevel {
	switch ll {
	case 0:
		return OnlyError
	case 1:
		return Verbose
	case 2:
		return Debug
	}
	return 1
}

func (l LogLevel) String() string {
	switch l {
	case OnlyError:
		return "OnlyErrors"
	case Verbose:
		return "Verbose"
	case Debug:
		return "Debug"
	}
	return "unknown"
}

type LoggerUtil struct {
	Level  LogLevel
	logger *log.Logger
}

func NewLoggerUtil(logLevel LogLevel) *LoggerUtil {
	return &LoggerUtil{Level: logLevel, logger: log.Default()}
}

func (l *LoggerUtil) LogPlain(formattedMessage string) {
	l.logger.Printf("%s ", formattedMessage)
}

func (l *LoggerUtil) LogError(formattedMessage string, err error) {
	l.logger.Printf("(ERROR) %s : %v", formattedMessage, err)
}

func (l *LoggerUtil) LogWhenVerbose(formattedMessage string) {
	if l.Level >= Verbose {
		l.logger.Printf("(VERBOSE) %s ", formattedMessage)
	}
}

func (l *LoggerUtil) LogWhenDebug(formattedMessage string) {
	if l.Level >= Debug {
		l.logger.Printf("%s (DEBUG) %s ", l.callerInfo(2), formattedMessage)
	}
}

func (l *LoggerUtil) LogIncomingRequest(request *http.Request) {
	if l.Level >= Debug {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			l.logger.Printf("error LogIncomingRequest: %v", err)
			return
		}
		requestStr := fmt.Sprintf("method: %s\nurl: '%s'\nbody: '%s'", request.Method, request.URL.String(), string(body))
		l.logger.Printf("%s (DEBUG) incoming request:\n%s", l.callerInfo(2), requestStr)
		request.Body = io.NopCloser(bytes.NewBuffer(body))
	}
}

func (l *LoggerUtil) callerInfo(skip int) (info string) {
	_, file, lineNo, _ := runtime.Caller(skip)
	fileName := path.Base(file)
	return fmt.Sprintf("%s:%d ", fileName, lineNo)
}

type LoggingResponseWriter struct {
	http.ResponseWriter
	responseCode int
	buf          *bytes.Buffer
	loggerUtil   *LoggerUtil
	skip         int
}

func NewLoggingResponseWriter(writer http.ResponseWriter, logger *LoggerUtil, skip int) *LoggingResponseWriter {
	lrw := &LoggingResponseWriter{
		ResponseWriter: writer,
		buf:            &bytes.Buffer{},
		loggerUtil:     logger,
		responseCode:   http.StatusOK,
		skip: skip,
	}
	return lrw
}

func (lrw *LoggingResponseWriter) Write(p []byte) (int, error) {
	return lrw.buf.Write(p)
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.responseCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Log() {
	if lrw.loggerUtil.Level >= Debug {
		lrw.loggerUtil.logger.Printf("%s (DEBUG) Sending response with status %d :\nbody:'%s'", lrw.loggerUtil.callerInfo(lrw.skip), lrw.responseCode, lrw.buf.String())
	}
	_, err := io.Copy(lrw.ResponseWriter, lrw.buf)
	if err != nil {
		lrw.loggerUtil.logger.Printf("LoggingResponseWriter: Failed to send out response: %v", err)
	}
}


