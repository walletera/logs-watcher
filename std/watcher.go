package std

import (
    "bufio"
    "fmt"
    "io"
    "os"

    "github.com/walletera/logs-watcher"
    "github.com/walletera/logs-watcher/newline"
)

// Watcher watches logs from Stdout and Stderr
type Watcher struct {
    *newline.Watcher

    realStdout *os.File
    realStderr *os.File
    pipeRead   *os.File
    pipeWrite  *os.File
    stopping   bool
}

// _ is a compile-time check ensuring that Watcher implements the logs.IWatcher interface.
var _ logs.IWatcher = (*Watcher)(nil)

func NewWatcher() (*Watcher, error) {
    stdoutWatcher := &Watcher{}

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

func (w *Watcher) startScanner(r io.Reader) {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        line := scanner.Text()
        _, err := fmt.Fprintln(w.realStdout, line)
        if err != nil {
            panic("failed writing log line to stdout: " + err.Error())
        }
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

func (w *Watcher) Stop() error {
    err := w.Watcher.Stop()
    if err != nil {
        return err
    }

    w.stopping = true

    os.Stdout = w.realStdout
    os.Stderr = w.realStderr

    err = w.pipeRead.Close()
    if err != nil {
        return fmt.Errorf("error stop logs watcher: %w", err)
    }
    err = w.pipeWrite.Close()
    if err != nil {
        return fmt.Errorf("error stop logs watcher: %w", err)
    }
    return nil
}
