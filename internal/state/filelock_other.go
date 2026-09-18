//go:build !unix

package state

func (l *fileLock) lock() error { return nil }

func (l *fileLock) unlock() error { return nil }
