package state

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	LockTimeout = 200 * time.Millisecond
	lockRetry   = 5 * time.Millisecond
)

var ErrBusy = errors.New("another writer holds the state file")

type fileLock struct {
	path string
	file *os.File
}

func newFileLock(path string) *fileLock {
	return &fileLock{path: path}
}

func (l *fileLock) Acquire() error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return fmt.Errorf("open state lock: %w", err)
	}
	l.file = f

	deadline := time.Now().Add(LockTimeout)
	for {
		err := l.tryLock()
		if err == nil {
			return nil
		}
		if !errors.Is(err, ErrBusy) {
			l.file = nil
			f.Close()
			return fmt.Errorf("lock state: %w", err)
		}
		if !time.Now().Before(deadline) {
			l.file = nil
			f.Close()
			return ErrBusy
		}
		time.Sleep(lockRetry)
	}
}

func (l *fileLock) Release() {
	if l.file == nil {
		return
	}
	l.unlock()
	l.file.Close()
	l.file = nil
}
