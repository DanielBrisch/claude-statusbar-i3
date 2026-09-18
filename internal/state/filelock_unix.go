//go:build unix

package state

import "syscall"

func (l *fileLock) lock() error {
	return syscall.Flock(int(l.file.Fd()), syscall.LOCK_EX)
}

func (l *fileLock) unlock() error {
	return syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
}
