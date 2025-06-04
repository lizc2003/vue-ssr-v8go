package tlog

import (
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

func Log(level LEVEL, file string, line string, msg string) {
	gLogger.pWithFileAndLine(level, file, line, msg)
}

func Debug(args ...interface{}) {
	gLogger.p(DEBUG, args...)
}

func Debugf(format string, args ...interface{}) {
	gLogger.pf(DEBUG, format, args...)
}

func Info(args ...interface{}) {
	gLogger.p(INFO, args...)
}

func Infof(format string, args ...interface{}) {
	gLogger.pf(INFO, format, args...)
}

func Warn(args ...interface{}) {
	gLogger.p(WARN, args...)
}

func Warnf(format string, args ...interface{}) {
	gLogger.pf(WARN, format, args...)
}

func Error(args ...interface{}) {
	gLogger.p(ERROR, args...)
}

func Errorf(format string, args ...interface{}) {
	gLogger.pf(ERROR, format, args...)
}

func Fatal(args ...interface{}) {
	gLogger.p(FATAL, args...)
	panic(fmt.Sprint(args...))
}

func Fatalf(format string, args ...interface{}) {
	gLogger.pf(FATAL, format, args...)
	panic(fmt.Sprintf(format, args...))
}
