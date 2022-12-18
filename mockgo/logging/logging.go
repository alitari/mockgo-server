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

// LogLevel level of log output
type LogLevel int64

const (
	// OnlyError logs just error cases, recommended for production use
	OnlyError LogLevel = iota
	// Verbose logs additional informations
	Verbose
	// Debug logs information for debugging use cases
	Debug
)

/*
ParseLogLevel parses the LogLevel from an integer value.
*/
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

/*
String gets the string representation of the LogLevel
*/
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

/*
LoggerUtil a helper for log output
*/
type LoggerUtil struct {
	Level  LogLevel
	logger *log.Logger
}

/*
NewLoggerUtil creates an instance of LoggerUtil
*/
func NewLoggerUtil(logLevel LogLevel) *LoggerUtil {
	return &LoggerUtil{Level: logLevel, logger: log.Default()}
}

/*
LogError logs an error case
*/
func (l *LoggerUtil) LogError(formattedMessage string, err error) {
	l.logger.Printf("(ERROR) %s : %v", formattedMessage, err)
}

/*
LogWhenVerbose logs an additional information
*/
func (l *LoggerUtil) LogWhenVerbose(formattedMessage string) {
	if l.Level >= Verbose {
		l.logger.Printf("(VERBOSE) %s ", formattedMessage)
	}
}

/*
LogWhenDebug logs a message for debugging purposes
*/
func (l *LoggerUtil) LogWhenDebug(formattedMessage string) {
	if l.Level >= Debug {
		l.logger.Printf("%s (DEBUG) %s ", l.callerInfo(2), formattedMessage)
	}
}

/*
LogIncomingRequest helper for logging http requests
*/
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

/*
ResponseWriter is a helper for logging http responses
*/
type ResponseWriter struct {
	http.ResponseWriter
	responseCode int
	buf          *bytes.Buffer
	loggerUtil   *LoggerUtil
	skip         int
}

/*
NewResponseWriter creates a new instance of ResponseWriter
*/
func NewResponseWriter(writer http.ResponseWriter, logger *LoggerUtil, skip int) *ResponseWriter {
	lrw := &ResponseWriter{
		ResponseWriter: writer,
		buf:            &bytes.Buffer{},
		loggerUtil:     logger,
		responseCode:   http.StatusOK,
		skip:           skip,
	}
	return lrw
}

func (lrw *ResponseWriter) Write(p []byte) (int, error) {
	return lrw.buf.Write(p)
}

// WriteHeader writes the response status code
func (lrw *ResponseWriter) WriteHeader(code int) {
	lrw.responseCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Log logs the response body
func (lrw *ResponseWriter) Log() {
	if lrw.loggerUtil.Level >= Debug {
		lrw.loggerUtil.logger.Printf("%s (DEBUG) Sending response with status %d :\nbody:'%s'", lrw.loggerUtil.callerInfo(lrw.skip), lrw.responseCode, lrw.buf.String())
	}
	_, err := io.Copy(lrw.ResponseWriter, lrw.buf)
	if err != nil {
		lrw.loggerUtil.logger.Printf("LoggingResponseWriter: Failed to send out response: %v", err)
	}
}
