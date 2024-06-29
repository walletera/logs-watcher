package std

import (
    "bufio"
    "fmt"
    "io"
    "os"

    "github.com/walletera/logs-watcher/newline"
)

// StdWatcher watches logs from Stdout and Stderr
type StdWatcher struct {
    *newline.Watcher

    realStdout *os.File
    realStderr *os.File
    pipeRead   *os.File
    pipeWrite  *os.File
    stopping   bool
}

func NewStdWatcher() (*StdWatcher, error) {
    stdoutWatcher := &StdWatcher{}

    stdoutWatcher.realStdout = os.Stdout
    stdoutWatcher.realStderr = os.Stderr

    r, w, err := os.Pipe()
    if err != nil {
        return nil, fmt.Errorf("failed starting logs watcher: %w", err)
    }

    stdoutWatcher.pipeRead = r
    stdoutWatcher.pipeWrite = w
    os.Stdout = w
    os.Stderr = w

    go stdoutWatcher.startScanner(r)

    baseWatcher := newline.NewWatcher()
    stdoutWatcher.Watcher = baseWatcher

    return stdoutWatcher, nil
}

func (w *StdWatcher) startScanner(r io.Reader) {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        line := scanner.Text()
        fmt.Fprintln(w.realStdout, line)
        w.Watcher.AddLogLine(line)
    }
    if err := scanner.Err(); err != nil {
        os.Stdout = w.realStdout
        os.Stderr = w.realStderr
        if !w.stopping {
            panic("failed reading log line from stdout or stderr: " + err.Error())
        }
    }
}

func (w *StdWatcher) Stop() error {
    w.Watcher.Stop()

    w.stopping = true

    os.Stdout = w.realStdout
    os.Stderr = w.realStderr

    err := w.pipeRead.Close()
    if err != nil {
        return fmt.Errorf("error stop logs watcher: %w", err)
    }
    err = w.pipeWrite.Close()
    if err != nil {
        return fmt.Errorf("error stop logs watcher: %w", err)
    }
    return nil
}
