//go:build !unix

package state

import "os"

func lockFile(*os.File) error { return nil }

func unlockFile(*os.File) error { return nil }
