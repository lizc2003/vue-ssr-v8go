package tlog

type LEVEL byte

const (
	DEBUG LEVEL = 1
	INFO  LEVEL = 2
	WARN  LEVEL = 3
	ERROR LEVEL = 4
	FATAL LEVEL = 5
)

var levelText = []string{"", "DEBUG", "INFO", "WARN", "ERROR", "FATAL", ""}

func getLevel(level string) LEVEL {
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}
