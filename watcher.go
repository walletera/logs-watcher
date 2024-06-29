package logs

import "time"

type Watcher interface {
    WaitFor(keyword string, timeout time.Duration) bool
    Stop() error
}
