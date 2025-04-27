package tlog

import (
	"context"
	"encoding/json"
)

type Fields map[string]any

type FieldsEntry struct {
	data Fields
}

func (e FieldsEntry) WithFields(fields Fields) FieldsEntry {
	data := e.data
	for k, v := range fields {
		data[k] = v
	}
	return FieldsEntry{data: data}
}

func (e FieldsEntry) WithField(k string, v any) FieldsEntry {
	data := e.data
	data[k] = v
	return FieldsEntry{data: data}
}

func (e FieldsEntry) makeOutFields() []field {
	sz := len(e.data)
	if sz == 0 {
		return nil
	}

	fields := make([]field, sz)
	i := 0
	for k, v := range e.data {
		bv, _ := json.Marshal(v)
		fields[i] = field{key: k, value: bv}
		i++
	}
	return fields
}

func (e FieldsEntry) Debug(args ...interface{}) {
	gLogger.p(DEBUG, e.makeOutFields(), args...)
}

func (e FieldsEntry) Debugf(format string, args ...interface{}) {
	gLogger.pf(DEBUG, e.makeOutFields(), format, args...)
}

func (e FieldsEntry) Info(args ...interface{}) {
	gLogger.p(INFO, e.makeOutFields(), args...)
}

func (e FieldsEntry) Infof(format string, args ...interface{}) {
	gLogger.pf(INFO, e.makeOutFields(), format, args...)
}

func (e FieldsEntry) Warn(args ...interface{}) {
	gLogger.p(WARN, e.makeOutFields(), args...)
}

func (e FieldsEntry) Warnf(format string, args ...interface{}) {
	gLogger.pf(WARN, e.makeOutFields(), format, args...)
}

func (e FieldsEntry) Error(args ...interface{}) {
	gLogger.p(ERROR, e.makeOutFields(), args...)
}

func (e FieldsEntry) Errorf(format string, args ...interface{}) {
	gLogger.pf(ERROR, e.makeOutFields(), format, args...)
}

func (e FieldsEntry) TraceDebug(ctx context.Context, format string, args ...any) {
	traceID, spanID := TraceAndSpanIdFromContext(ctx)
	gLogger.pTrace(DEBUG, traceID, spanID, e.makeOutFields(), format, args...)
}

func (e FieldsEntry) TraceInfo(ctx context.Context, format string, args ...any) {
	traceID, spanID := TraceAndSpanIdFromContext(ctx)
	gLogger.pTrace(INFO, traceID, spanID, e.makeOutFields(), format, args...)
}

func (e FieldsEntry) TraceWarn(ctx context.Context, format string, args ...any) {
	traceID, spanID := TraceAndSpanIdFromContext(ctx)
	gLogger.pTrace(WARN, traceID, spanID, e.makeOutFields(), format, args...)
}

func (e FieldsEntry) TraceError(ctx context.Context, format string, args ...any) {
	traceID, spanID := TraceAndSpanIdFromContext(ctx)
	gLogger.pTrace(ERROR, traceID, spanID, e.makeOutFields(), format, args...)
}
