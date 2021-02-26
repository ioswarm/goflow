package goflow

import "time"

type LogType int

const (
	DEBUG LogType = iota
	INFO
	WARNING
	ERROR
	FATAL
	PANIC
)

var LogTypeNames = []string{"DEBUG", "INFO", "WARNING", "ERROR", "FATAL", "PANIC", "UNKNOWN"}

func LogTypeName(i int) string {
	l := len(LogTypeNames)
	if i < 0 || i >= l {
		return LogTypeNames[l-1]
	}
	return LogTypeNames[i]
}

type Logger interface {
	DEBUG(msg string, a ...interface{})
	INFO(msg string, a ...interface{})
	WARN(msg string, a ...interface{})
	ERROR(msg string, a ...interface{})
	FATAL(msg string, a ...interface{})
	PANIC(msg string, a ...interface{})
}

type LogEntry struct {
	LogType   LogType
	Timestamp time.Time
	Message   string
}

func (le *LogEntry) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"logType": LogTypeName(int(le.LogType)),
		"timestamp": le.Timestamp.UTC().Format("2006-01-02T15:04:05.999Z"),
		"message": le.Message,
	}
}
