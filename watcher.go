package logs

import "time"

type Watcher interface {
    WaitForNTimes(keyword string, timeout time.Duration, n int) bool
    WaitFor(keyword string, timeout time.Duration) bool
    Stop() error
}
