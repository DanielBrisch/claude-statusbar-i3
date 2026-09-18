package state

import (
	"fmt"
	"os"
)

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
	if err := l.lock(); err != nil {
		l.file = nil
		f.Close()
		return fmt.Errorf("lock state: %w", err)
	}
	return nil
}

func (l *fileLock) Release() {
	if l.file == nil {
		return
	}
	l.unlock()
	l.file.Close()
	l.file = nil
}
