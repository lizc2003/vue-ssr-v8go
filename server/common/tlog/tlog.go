package tlog

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	FileSize int    `toml:"filesize" json:"filesize"`
	FileNum  int    `toml:"filenum" json:"filenum"`
	FileName string `toml:"filename" json:"filename"`
	Level    string `toml:"level" json:"level"`
	Debug    bool   `toml:"debug" json:"debug"`
	Dir      string `toml:"dir" json:"dir"`
	UseJson  bool   `toml:"usejson" json:"usejson"`
}

var stdLogger = newLogger(&Config{Debug: true}, defaultServerName)
var gLogger = stdLogger

func TraceAndSpanIdFromContext(ctx context.Context) (string, string) {
	return "", ""
}

func (c *Config) check(serverName string, suffix string) {
	if c.FileSize == 0 {
		c.FileSize = 128
	}
	if c.FileNum == 0 {
		c.FileNum = 10
	}
	if c.FileName == "" {
		c.FileName = "INFO"
	}
	if c.Level == "" {
		c.Level = "DEBUG"
	}

	if c.Dir == "" {
		c.Dir = "./logs"
	} else {
		const svrNameTag = "{svr.name}"
		fileDir := strings.Trim(c.Dir, " ")
		fileDir = strings.ReplaceAll(fileDir, svrNameTag, serverName)
		if suffix != "" {
			fileDir += "." + suffix
		}
		if strings.HasPrefix(fileDir, ".") {
			execPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			fileDir = path.Join(execPath, fileDir)
		}
		c.Dir = fileDir
	}
}

func Init(c Config, serverName string, suffix string) {
	if gLogger == stdLogger {
		c.check(serverName, suffix)
		l := newLogger(&c, serverName)
		if l != nil {
			gLogger = l
		} else {
			fmt.Println("init logger failed")
		}
	}
}

func Close() {
	if gLogger != stdLogger {
		tmp := gLogger
		gLogger = stdLogger
		time.Sleep(100 * time.Millisecond)
		tmp.stop()
	} else {
		time.Sleep(100 * time.Millisecond)
	}
}

func IsDebugEnabled() bool {
	return gLogger.level <= DEBUG
}

func Level() LEVEL {
	return gLogger.getLevel()
}

func Log(level LEVEL, args ...interface{}) {
	gLogger.p(level, nil, args...)
}

func Logf(level LEVEL, format string, args ...interface{}) {
	gLogger.pf(level, nil, format, args...)
}

func Debug(args ...interface{}) {
	gLogger.p(DEBUG, nil, args...)
}

func Debugf(format string, args ...interface{}) {
	gLogger.pf(DEBUG, nil, format, args...)
}

func Info(args ...interface{}) {
	gLogger.p(INFO, nil, args...)
}

func Infof(format string, args ...interface{}) {
	gLogger.pf(INFO, nil, format, args...)
}

func Warn(args ...interface{}) {
	gLogger.p(WARN, nil, args...)
}

func Warnf(format string, args ...interface{}) {
	gLogger.pf(WARN, nil, format, args...)
}

func Error(args ...interface{}) {
	gLogger.p(ERROR, nil, args...)
}

func Errorf(format string, args ...interface{}) {
	gLogger.pf(ERROR, nil, format, args...)
}

func Fatal(args ...interface{}) {
	gLogger.p(FATAL, nil, args...)
	panic(fmt.Sprint(args...))
}

func Fatalf(format string, args ...interface{}) {
	gLogger.pf(FATAL, nil, format, args...)
	panic(fmt.Sprintf(format, args...))
}

func TraceDebug(ctx context.Context, format string, args ...any) {
	traceID, spanID := TraceAndSpanIdFromContext(ctx)
	gLogger.pTrace(DEBUG, traceID, spanID, nil, format, args...)
}

func TraceInfo(ctx context.Context, format string, args ...any) {
	traceID, spanID := TraceAndSpanIdFromContext(ctx)
	gLogger.pTrace(INFO, traceID, spanID, nil, format, args...)
}

func TraceWarn(ctx context.Context, format string, args ...any) {
	traceID, spanID := TraceAndSpanIdFromContext(ctx)
	gLogger.pTrace(WARN, traceID, spanID, nil, format, args...)
}

func TraceError(ctx context.Context, format string, args ...any) {
	traceID, spanID := TraceAndSpanIdFromContext(ctx)
	gLogger.pTrace(ERROR, traceID, spanID, nil, format, args...)
}

func WithFields(fields Fields) FieldsEntry {
	return FieldsEntry{data: fields}
}

func WithField(k string, v any) FieldsEntry {
	return FieldsEntry{data: Fields{k: v}}
}

type IoWriter struct{}

func (io *IoWriter) Write(p []byte) (n int, err error) {
	gLogger.p(INFO, nil, "gin log:", string(p))
	return len(p), nil
}
