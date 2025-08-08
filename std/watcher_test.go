package std

import (
    "fmt"
    "math/rand"
    "sync"
    "testing"
    "time"

    "github.com/walletera/logs-watcher/testdata"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestStdWatcher_WaitFor_LogIsAlreadyThere(t *testing.T) {
    logsWatcher, err := NewWatcher()
    require.NoError(t, err)

    time.Sleep(1 * time.Millisecond)

    fmt.Println("Hola Mundo Loco!")

    found := logsWatcher.WaitFor("Mundo", 100*time.Millisecond)

    assert.True(t, found)

    err = logsWatcher.Stop()
    require.NoError(t, err)
}

func TestStdWatcher_WaitFor_LogAppearsAfterTheCallToWaitFor(t *testing.T) {
    logsWatcher, err := NewWatcher()
    require.NoError(t, err)

    fmt.Println("Hola")

    go func() {
        time.Sleep(100 * time.Millisecond)
        fmt.Println("Mundo Loco!")
    }()

    found := logsWatcher.WaitFor("Mundo", 200*time.Millisecond)

    assert.True(t, found)

    err = logsWatcher.Stop()
    require.NoError(t, err)
}

func TestStdWatcher_WaitFor_LogAppearsTooLate(t *testing.T) {
    logsWatcher, err := NewWatcher()
    require.NoError(t, err)

    fmt.Println("Hola")

    go func() {
        time.Sleep(200 * time.Millisecond)
        fmt.Println("Mundo Loco!")
    }()

    found := logsWatcher.WaitFor("Mundo", 100*time.Millisecond)
    assert.False(t, found)

    // Let the log appear in the console
    time.Sleep(100 * time.Millisecond)
    err = logsWatcher.Stop()
    require.NoError(t, err)
}

func TestStdWatcher_WaitFor_MultilineLog(t *testing.T) {
    logsWatcher, err := NewWatcher()
    require.NoError(t, err)

    go func() {
        time.Sleep(100 * time.Millisecond)
        fmt.Print(testdata.MultilineLog)
    }()

    found := logsWatcher.WaitFor("failed creating payment on dinopay", 200*time.Millisecond)

    assert.True(t, found)

    err = logsWatcher.Stop()
    require.NoError(t, err)
}

func TestStdWatcher_WaitFor_Concurrency(t *testing.T) {
    if len(testdata.Syslog) != len(testdata.SyslogSubstrs) {
        t.Error("Syslog and SyslogSubstrs are not the same length")
    }

    logsWatcher, err := NewWatcher()
    require.NoError(t, err)

    goroutinesCount := 100
    wg := &sync.WaitGroup{}
    wg.Add(goroutinesCount)
    syslogIndexCh := make(chan int, goroutinesCount)
    syslogSubstrIndexCh := make(chan int, goroutinesCount)

    for i := 0; i < goroutinesCount; i++ {
        go func() {
            index := <-syslogIndexCh
            sysLogEntry := testdata.Syslog[index]
            fmt.Println(sysLogEntry)
        }()
    }

    for i := 0; i < goroutinesCount; i++ {
        go func() {
            index := <-syslogSubstrIndexCh
            keyword := testdata.SyslogSubstrs[index]
            found := logsWatcher.WaitFor(keyword, 1*time.Second)
            assert.True(t, found, "keyword not found: ", keyword)
            wg.Done()
        }()
    }

    for i := 0; i < goroutinesCount; i++ {
        randomIndex := rand.Intn(len(testdata.SyslogSubstrs))
        go func() {
            syslogIndexCh <- randomIndex
        }()
        go func() {
            syslogSubstrIndexCh <- randomIndex
        }()
    }

    wg.Wait()

    err = logsWatcher.Stop()
    require.NoError(t, err)
}
