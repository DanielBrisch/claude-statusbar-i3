//go:build !unix

package state

func (l *fileLock) tryLock() error { return nil }

func (l *fileLock) unlock() error { return nil }
