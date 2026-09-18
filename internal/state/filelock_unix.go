//go:build unix

package state

import (
	"errors"
	"syscall"
)

func (l *fileLock) tryLock() error {
	err := syscall.Flock(int(l.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return ErrBusy
	}
	return err
}

func (l *fileLock) unlock() error {
	return syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
}
