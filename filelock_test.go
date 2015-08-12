package filelock

import (
  "sync"
  "testing"
  "time"
)

func TestRepeatedObtainAndReleaseLock(t *testing.T) {
  count := 100

  for i := 0; i < count; i++ {
    l := Obtain("test.lock", time.Second * 1)

    if l.State != LockSuccess {
      t.Error()
    }

    if l.Release(); l.State != LockReleased {
      t.Error()
    }
  }  
}

func TestObtainAndReleaseLockConcurrent(t *testing.T) {
  count := 50
  var wg sync.WaitGroup
  wg.Add(count)

  startTime := time.Now()
  lockTime := time.Millisecond * 5

  for i := 0; i < count; i++ {
    go func() {
      l := Obtain(".test1.lock", time.Second * 1)
      if l.State != LockSuccess {
        t.Error()
      }

      time.Sleep(lockTime)

      if l.Release(); l.State != LockReleased {
        t.Error()
      }

      wg.Done()      
    }()
  }

  wg.Wait()

  // Each goroutine held the lock for lockTime, so the
  // test duration should be at least lockTime * count
  duration := time.Since(startTime)
  if int(duration / lockTime) < count {
    t.Error()
  }
}

func TestObtainLockTimeout(t *testing.T) {
  l := Obtain(".test2.lock", time.Second * 1)
  if l.State != LockSuccess {
    t.Error()
  }

  defer l.Release()

  if l2 := Obtain(".test2.lock", time.Millisecond * 10); l2.State != LockTimeout {
    t.Error()
  }
}

func TestObtainLockTimeoutReleasesEventuallyObtainedLock(t *testing.T) {
  l1 := Obtain(".test3.lock", time.Second * 1)

  if l1.State != LockSuccess {
    t.Error()
  }

  l2 := Obtain(".test3.lock", time.Millisecond * 10)

  if l2.State != LockTimeout {
    t.Error()
  }

  // Release the first lock, this causes the
  // second lock to be obtained by the blocking goroutine,
  // and then be released straight away.  
  l1.Release()

  // Given that the first lock has been released, and
  // the second lock timed out and should have been released
  // as soon as it was obtained by the still-blocking goroutine,
  // a new lock should succeed.
  l3 := Obtain(".test3.lock", time.Millisecond * 10)

  if l3.State != LockSuccess {
    t.Error()
  }
}

func TestLockFilePermissionDenied(t *testing.T) {
  l := Obtain("/.test.lock", time.Second * 1)

  if l.State != LockError {
    t.Error()
  }
}
