package filelock

import (
  "os"
  "time"
  "syscall"
)

type Lock struct {
  State int
  Error error

  timeout time.Duration
  path string
  file *os.File
}

const (
  LockSuccess int = 1 << iota
  LockError
  LockTimeout
  LockReleased
)

func Obtain(path string, timeout time.Duration) (lock *Lock) {
  lock = &Lock{path: path, timeout: timeout}

  // Try to open the lock file
  file, err := os.OpenFile(lock.path, os.O_RDWR | os.O_CREATE, 0660)

  if err != nil {
    lock.Error = err
    lock.State = LockError
    return
  }

  // Create channels for the timeout and receiving the flock result
  timeoutChan := time.After(lock.timeout)
  flockChan := make(chan error, 1)

  // Start the blocking lock call in a goroutine
  go func() { flockChan <- flock(file) }()

  select {
    case <- timeoutChan:
      lock.State = LockTimeout

      // We hit the timeout without successfully getting the lock.
      // The goroutine blocked on syscall.Flock() is still running
      // and will eventually return at some point in the future.
      // If the lock is eventually obtained, it needs to be released.
      go func() {
        if err := <- flockChan; err == nil {
          releaseFlock(file)
        }
      }()

    case err := <- flockChan:
      if err == nil {
        lock.State = LockSuccess
        lock.file = file
      } else {
        lock.State = LockError
        lock.Error = err
      }
  }

  return
}

func (l *Lock) IsLocked() bool {
  return l.State == LockSuccess
}

func (l *Lock) Release() {
  if l.State == LockSuccess {
    releaseFlock(l.file)
    l.file.Close()
    l.file = nil
    l.State = LockReleased
  }
}

func flock(file *os.File) error {
  return syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
}

func releaseFlock(file *os.File) {
  syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
