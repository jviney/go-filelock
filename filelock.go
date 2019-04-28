package filelock

import (
	"errors"
	"os"
	"syscall"
	"time"
)

var ErrLockTimeout = errors.New("timeout obtaining lock")
var ErrNotLocked = errors.New("not locked")

type FileLock struct {
	Path    string
	Timeout time.Duration

	file *os.File // open file holding the lock
}

func (l *FileLock) Lock() error {
	// Try to open the lock file
	file, err := os.OpenFile(l.Path, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return err
	}

	// Create channels for the timeout and receiving the flock result
	timeoutChan := time.After(l.Timeout)
	flockChan := make(chan error, 1)

	// Start the blocking flock call in a goroutine
	go func() { flockChan <- flock(file) }()

	select {
	case <-timeoutChan:
		// We hit the timeout without successfully getting the lock.
		// The goroutine blocked on syscall.Flock() is still running
		// and will eventually return at some point in the future.
		// When the lock is eventually obtained, it needs to be immediately
		// released.
		go func() {
			if err := <-flockChan; err == nil {
				releaseFlock(file)
				l.file = nil
			}
		}()

		l.file = nil
		return ErrLockTimeout

	case err := <-flockChan:
		if err != nil {
			return err
		}

		// Store the file descriptor holding the lock
		l.file = file
		return nil
	}
}

func (l *FileLock) Unlock() error {
	if l.file == nil {
		return ErrNotLocked
	}

	err := releaseFlock(l.file)
	if err != nil {
		return err
	}

	err = l.file.Close()
	if err != nil {
		return err
	}

	l.file = nil
	return nil
}

func flock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
}

func releaseFlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
