package slog

import (
    "fmt"
    "log/slog"
    "regexp"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type simpleWriter struct {
    content []byte
}

func (w *simpleWriter) Write(p []byte) (n int, err error) {
    w.content = p
    return len(p), nil
}

func TestHandlerDecorator_WithAttrs(t *testing.T) {
    writer := &simpleWriter{}
    decorator := NewHandlerDecorator(slog.NewTextHandler(writer, nil), func(record slog.Record) {})
    logger := slog.New(decorator)
    logger1 := logger.With("attribute1", "value1")
    logger.Info("test message")
    matched, err := regexp.Match("attribute1=value1", writer.content)
    require.NoError(t, err)
    fmt.Printf("%s\n", writer.content)
    assert.False(t, matched, "attribute1 should not be present in logger logs")
    logger1.Info("test message")
    fmt.Printf("%s\n", writer.content)
    assert.False(t, matched, "attribute1 should be present in logger1 logs")
}
