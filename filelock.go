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

func Obtain(path string, timeout time.Duration) *Lock {
  lock := Lock{path: path, timeout: timeout}
  timeoutChan := time.After(lock.timeout)

  flockChan := make(chan bool, 1)
  go func() { flockChan <- lock.flock() }()

  select {
    case <- timeoutChan:
      lock.State = LockTimeout

      // We hit the timeout without successfully getting the lock.
      // The goroutine blocked on syscall.Flock() is still running
      // and will eventually return at some point in the future.
      // If the lock is eventually obtained, it needs to be released.
      go func() {
        if <- flockChan {
          lock.releaseFlock()
        }
      }()

    case success := <- flockChan:
      if success {
        lock.State = LockSuccess
      } else {
        lock.State = LockError
      }
  }
  return &lock
}

func (l *Lock) IsLocked() bool {
  return l.State == LockSuccess
}

func (l *Lock) Release() {
  if l.State == LockSuccess {
    l.releaseFlock()
    l.State = LockReleased
  }
}

func (l *Lock) flock() bool {
  file, err := os.OpenFile(l.path, os.O_RDWR | os.O_CREATE, 0660)

  if err != nil {
    l.Error = err
    return false
  }

  if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
    l.Error = err
    return false
  } else {
    l.file = file
    return true
  }
}

func (l *Lock) releaseFlock() {
  if l.file != nil {
    syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
    l.file.Close()
    l.file = nil
  } 
}
