package slog

import (
    "log/slog"
    "math/rand"
    "sync"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/walletera/logs-watcher/testdata"
    "go.uber.org/zap"
    "go.uber.org/zap/exp/zapslog"
)

func TestSlogWatcher_WaitFor_LogIsAlreadyThere(t *testing.T) {
    handler, err := newZapHandler()
    require.NoError(t, err)

    logsWatcher := NewWatcher(handler)
    logger := slog.New(logsWatcher.decoratedHandler)

    time.Sleep(1 * time.Millisecond)

    logger.Info("Hola Mundo Loco!")

    found := logsWatcher.WaitFor("Mundo", 100*time.Millisecond)

    assert.True(t, found)

    err = logsWatcher.Stop()
    require.NoError(t, err)
}

func TestSlogWatcher_WaitFor_LogAppearsAfterTheCallToWaitFor(t *testing.T) {
    handler, err := newZapHandler()
    require.NoError(t, err)

    logsWatcher := NewWatcher(handler)
    logger := slog.New(logsWatcher.decoratedHandler)

    logger.Info("Hola")

    go func() {
        time.Sleep(100 * time.Millisecond)
        logger.Info("Mundo Loco!")
    }()

    found := logsWatcher.WaitFor("Mundo", 200*time.Millisecond)

    assert.True(t, found)

    err = logsWatcher.Stop()
    require.NoError(t, err)
}

func TestSlogWatcher_WaitFor_LogAppearsTooLate(t *testing.T) {
    handler, err := newZapHandler()
    require.NoError(t, err)

    logsWatcher := NewWatcher(handler)
    logger := slog.New(logsWatcher.decoratedHandler)

    logger.Info("Hola")

    go func() {
        time.Sleep(200 * time.Millisecond)
        logger.Info("Mundo Loco!")
    }()

    found := logsWatcher.WaitFor("Mundo", 100*time.Millisecond)
    assert.False(t, found)

    // Let the log appear in the console
    time.Sleep(100 * time.Millisecond)
    err = logsWatcher.Stop()
    require.NoError(t, err)
}

func TestSlogWatcher_WaitFor_MultilineLog(t *testing.T) {
    handler, err := newZapHandler()
    require.NoError(t, err)

    logsWatcher := NewWatcher(handler)
    logger := slog.New(logsWatcher.decoratedHandler)

    go func() {
        time.Sleep(100 * time.Millisecond)
        logger.Info(testdata.MultilineLog)
    }()

    found := logsWatcher.WaitFor("failed creating payment on dinopay", 200*time.Millisecond)

    assert.True(t, found)

    err = logsWatcher.Stop()
    require.NoError(t, err)
}

func TestSlogWatcher_WaitFor_Concurrency(t *testing.T) {
    handler, err := newZapHandler()
    require.NoError(t, err)

    logsWatcher := NewWatcher(handler)
    logger := slog.New(logsWatcher.decoratedHandler)

    goroutinesCount := 100

    for i := 0; i < goroutinesCount; i++ {
        go func() {
            sysLogEntry := testdata.Syslog[rand.Intn(len(testdata.Syslog))]
            logger.Info(sysLogEntry)
        }()
    }

    wg := &sync.WaitGroup{}
    wg.Add(goroutinesCount)

    for i := 0; i < goroutinesCount; i++ {
        go func() {
            keyword := testdata.SyslogSubstrs[rand.Intn(len(testdata.SyslogSubstrs))]
            found := logsWatcher.WaitFor(keyword, time.Duration(goroutinesCount+1)*time.Millisecond)
            assert.True(t, found, "keyword not found: ", keyword)
            wg.Done()
        }()
    }

    wg.Wait()

    err = logsWatcher.Stop()
    require.NoError(t, err)
}

func newZapHandler() (slog.Handler, error) {
    zapConfig := zap.Config{
        Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
        Development: false,
        Sampling: &zap.SamplingConfig{
            Initial:    100,
            Thereafter: 100,
        },
        Encoding:         "json",
        EncoderConfig:    zap.NewProductionEncoderConfig(),
        OutputPaths:      []string{"stderr"},
        ErrorOutputPaths: []string{"stderr"},
    }
    zapLogger, err := zapConfig.Build()
    if err != nil {
        return nil, err
    }
    return zapslog.NewHandler(zapLogger.Core(), nil), nil
}
