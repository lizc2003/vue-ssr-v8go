package util

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unsafe"
)

func FormatTime(t time.Time) string {
	y, m, d := t.Date()
	hour, minute, second := t.Clock()
	return fmt.Sprintf("%4d-%02d-%02d %02d:%02d:%02d", y, m, d, hour, minute, second)
}

func ParseTime(s string) (time.Time, error) {
	if strings.Index(s, ",") >= 0 {
		return time.Parse(time.RFC1123, s)
	} else if strings.Index(s, "T") >= 0 {
		if len(s) == 19 {
			return time.Time{}, errors.New("inconsistent with the rfc3339")
		}
		return time.Parse(time.RFC3339, s)
	} else if strings.Index(s, ":") >= 0 {
		return time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	} else {
		return time.ParseInLocation("2006-01-02", s, time.Local)
	}
}

func UnsafeStr2Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// []byte转string
func UnsafeBytes2Str(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
