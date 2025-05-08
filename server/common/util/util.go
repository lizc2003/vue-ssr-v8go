package util

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	"unsafe"
)

func FormatTime(t time.Time) string {
	y, m, d := t.Date()
	hour, minute, second := t.Clock()
	return fmt.Sprintf("%4d-%02d-%02d %02d:%02d:%02d", y, m, d, hour, minute, second)
}

func FileExists(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err == nil {
		return !info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func GetClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				return ip
			}
		}
	}

	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	remoteAddr := r.RemoteAddr
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return ip
}

func WriteHtmlResponse(w http.ResponseWriter, status int, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	w.Write(UnsafeStr2Bytes(html))
}

func UnsafeStr2Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func UnsafeBytes2Str(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
