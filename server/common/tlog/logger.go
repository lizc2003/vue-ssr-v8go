package tlog

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Logger struct {
	serverName string
	fileSize   int64
	fileNum    int
	fileName   string
	dir        string
	host       string
	level      LEVEL
	timeOffset []byte
	byteBuff   bytes.Buffer
	queue      chan *Msg
	f          *os.File
	w          io.Writer
	ticker     *time.Ticker
	end        chan struct{}
	useJson    bool
}

type Msg struct {
	line  string
	file  string
	level LEVEL
	msg   []byte
}

type field struct {
	key   string
	value []byte
}

const (
	defaultServerName = "default"
)

func newLogger(c *Config, serverName string) *Logger {
	var fName string
	var f *os.File
	var w io.Writer
	var ticker *time.Ticker
	var timeOffset []byte

	level := getLevel(c.Level)
	queueSize := 102400
	host, _ := os.Hostname()

	_, offset := time.Now().Zone()
	offset /= 60
	if offset == 0 {
		timeOffset = []byte("Z")
	} else if offset < 0 {
		offset = -offset
		timeOffset = []byte(fmt.Sprintf("-%02d:%02d", offset/60, offset%60))
	} else {
		timeOffset = []byte(fmt.Sprintf("+%02d:%02d", offset/60, offset%60))
	}

	if c.Debug {
		f = os.Stdout
		w = f
		level = DEBUG
		if serverName == defaultServerName {
			queueSize = 1024
		}
	} else {
		var err error
		os.MkdirAll(c.Dir, 0755)
		fName = path.Join(c.Dir, c.FileName+".log")
		f, err = os.OpenFile(fName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		w = bufio.NewWriterSize(f, 1024*1024)
		ticker = time.NewTicker(2 * time.Second)
	}

	l := &Logger{
		serverName: serverName,
		fileSize:   int64(c.FileSize * 1024 * 1024),
		fileNum:    c.FileNum,
		fileName:   fName,
		dir:        c.Dir,
		host:       host,
		level:      level,
		timeOffset: timeOffset,
		queue:      make(chan *Msg, queueSize),
		f:          f,
		w:          w,
		ticker:     ticker,
		end:        make(chan struct{}),
		useJson:    c.UseJson,
	}

	go l.writeLoop()
	if l.ticker != nil {
		go l.flushLoop()
	}
	return l
}

func (l *Logger) getLevel() LEVEL {
	return l.level
}

func (l *Logger) stop() {
	if l.ticker != nil {
		l.ticker.Stop()
	}
	close(l.queue)
	<-l.end

	if l.w != nil {
		if bufW, ok := l.w.(*bufio.Writer); ok {
			bufW.Flush()
		}
	}
	if l.f != nil && l.f != os.Stdout {
		l.f.Close()
	}
}

func (l *Logger) writeLoop() {
	for a := range l.queue {
		if a == nil {
			if bufW, ok := l.w.(*bufio.Writer); ok {
				bufW.Flush()
				fileInfo, err := os.Stat(l.fileName)
				if err != nil {
					if os.IsNotExist(err) {
						l.f.Close()
						os.MkdirAll(l.dir, 0755)
						l.f, _ = os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						bufW.Reset(l.f)
					}
				} else if fileInfo.Size() > l.fileSize {
					l.f.Close()
					os.Rename(l.fileName, l.makeOldName())
					l.f, _ = os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					bufW.Reset(l.f)
					l.rmOldFiles()
				}
			}
		} else {
			l.w.Write(l.makeLog(&l.byteBuff, a))
			l.byteBuff.Reset()
		}
	}

	close(l.end)
}

func (l *Logger) flushLoop() {
	for range l.ticker.C {
		l.queue <- nil
	}
}

func (l *Logger) makeOldName() string {
	t := fmt.Sprintf("%s", time.Now())[:19]
	tt := strings.Replace(
		strings.Replace(
			strings.Replace(t, "-", "", -1),
			" ", "", -1),
		":", "", -1)
	return fmt.Sprintf("%s.%s", l.fileName, tt)
}

func (l *Logger) pWithFileAndLine(level LEVEL, file string, line string, msg string) {
	if level >= l.level {
		b := []byte(msg)
		m := &Msg{file: file, line: line, level: level, msg: b}

		select {
		case l.queue <- m:
		default:
		}
	}
}

func (l *Logger) p(level LEVEL, args ...interface{}) {
	if level >= l.level {
		file, line := getFileNameAndLine()
		var w bytes.Buffer
		for _, arg := range args {
			fmt.Fprint(&w, arg)
			w.WriteByte(' ')
		}
		b := w.Bytes()
		m := &Msg{file: file, line: line, level: level, msg: b}

		select {
		case l.queue <- m:
		default:
		}
	}
}

func (l *Logger) pf(level LEVEL, format string, args ...interface{}) {
	if level >= l.level {
		file, line := getFileNameAndLine()
		var w bytes.Buffer
		fmt.Fprintf(&w, format, args...)
		b := w.Bytes()
		m := &Msg{file: file, line: line, level: level, msg: b}

		select {
		case l.queue <- m:
		default:
		}
	}
}

func (l *Logger) makeLog(w *bytes.Buffer, msg *Msg) []byte {
	if l.useJson {
		return l.makeJsonLog(w, msg)
	}

	w.Write(l.genTime())
	w.WriteByte(' ')
	w.WriteString(l.host)
	w.WriteByte(' ')
	w.WriteString(l.serverName)
	w.WriteByte(' ')
	w.WriteString(levelText[msg.level])
	w.WriteByte(' ')
	w.WriteString(msg.file)
	w.WriteByte(':')
	w.WriteString(msg.line)
	w.WriteByte(' ')

	if l.level > DEBUG {
		w.Write(bytes.ReplaceAll(msg.msg, []byte{'\n'}, []byte{' '}))
	} else {
		w.Write(msg.msg)
	}
	w.WriteByte('\n')
	return w.Bytes()
}

func (l *Logger) rmOldFiles() {
	entries, err := os.ReadDir(l.dir)
	if err != nil {
		return
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	idx := 0
	for i := len(names) - 1; i >= 0; i-- {
		file := path.Join(l.dir, names[i])
		if strings.HasPrefix(file, l.fileName) && file != l.fileName {
			idx++
			if idx > l.fileNum {
				_ = os.Remove(file)
			}
		}
	}
}

func (l *Logger) genTime() []byte {
	now := time.Now()
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	return append([]byte{
		'2', byte((year%1000)/100) + 48, byte((year%100)/10) + 48, byte(year%10) + 48, '-',
		byte(month/10) + 48, byte(month%10) + 48, '-',
		byte(day/10) + 48, byte(day%10) + 48, 'T',
		byte(hour/10) + 48, byte(hour%10) + 48, ':',
		byte(minute/10) + 48, byte(minute%10) + 48, ':',
		byte(second/10) + 48, byte(second%10) + 48}, l.timeOffset...)
}

func getFileNameAndLine() (string, string) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "???", "0"
	}
	dirs := strings.Split(file, "/")
	sz := len(dirs)
	if sz >= 2 {
		file = dirs[sz-2] + "/" + dirs[sz-1]
	}
	return file, strconv.FormatInt(int64(line), 10)
}
