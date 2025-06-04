package tlog

import (
	"bytes"
	"encoding/json"
	"unsafe"
)

func (l *Logger) makeJsonLog(w *bytes.Buffer, msg *Msg) []byte {
	w.WriteByte('{')

	w.WriteString("\"time\":")
	w.WriteByte('"')
	w.Write(l.genTime())
	w.WriteByte('"')

	w.WriteString(",\"host\":")
	w.WriteByte('"')
	w.WriteString(l.host)
	w.WriteByte('"')

	w.WriteString(",\"server\":")
	w.WriteByte('"')
	w.WriteString(l.serverName)
	w.WriteByte('"')

	w.WriteString(",\"level\":")
	w.WriteByte('"')
	w.WriteString(levelText[msg.level])
	w.WriteByte('"')

	w.WriteString(",\"file\":")
	w.WriteByte('"')
	w.WriteString(msg.file)
	w.WriteByte('"')

	w.WriteString(",\"line\":")
	w.WriteByte('"')
	w.WriteString(msg.line)
	w.WriteByte('"')

	if len(msg.msg) > 0 {
		w.WriteString(",\"msg\":")
		b, _ := json.Marshal(unsafeBytes2Str(msg.msg))
		w.Write(b)
	}

	w.WriteString("}\n")
	return w.Bytes()
}

func unsafeBytes2Str(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func unsafeStr2Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
