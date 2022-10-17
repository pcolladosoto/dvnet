package dvnet

import lg "log"

type logLevel uint

const (
	LogLevelDebug logLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelErr
)

type logger struct {
	level logLevel
}

var log logger

func (l *logger) debug(fmt string, args ...interface{}) {
	if l.level < LogLevelInfo {
		lg.Printf("DEBUG: "+fmt, args...)
	}
}

func (l *logger) info(fmt string, args ...interface{}) {
	if l.level < LogLevelInfo {
		lg.Printf("INFO: "+fmt, args...)
	}
}

func (l *logger) warn(fmt string, args ...interface{}) {
	lg.Printf("WARNING: "+fmt, args...)
}

func (l *logger) error(fmt string, args ...interface{}) {
	lg.Printf("ERROR: "+fmt, args...)
}

func InitLogger(level logLevel) {
	lg.SetFlags(lg.Flags() & ^(lg.Lmicroseconds | lg.Ldate))
	log = logger{level}
}
