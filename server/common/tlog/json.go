package tlog

import (
	"bytes"
	"encoding/json"
	"strconv"
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
	w.WriteString(strconv.FormatInt(int64(msg.line), 10))
	w.WriteByte('"')

	if len(msg.traceId) > 0 {
		w.WriteString(",\"trace_id\":")
		w.WriteByte('"')
		w.WriteString(msg.traceId)
		w.WriteByte('"')

		if len(msg.spanId) > 0 {
			w.WriteString(",\"span_id\":")
			w.WriteByte('"')
			w.WriteString(msg.spanId)
			w.WriteByte('"')
		}
	}

	for _, f := range msg.fields {
		w.WriteString(",\"")
		w.WriteString(f.key)
		w.WriteString("\":")
		w.Write(f.value)
	}

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
