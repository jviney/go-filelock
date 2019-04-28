package filelock

import (
	"sync"
	"testing"
	"time"
)

func TestRepeatedObtainAndReleaseLock(t *testing.T) {
	count := 100

	for i := 0; i < count; i++ {
		lock := FileLock{Path: "/tmp/.test.lock", Timeout: time.Second * 1}
		err := lock.Lock()
		if err != nil {
			t.Error()
		}

		if err = lock.Unlock(); err != nil {
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
			lock := FileLock{Path: "/tmp/.test1.lock", Timeout: time.Second * 1}
			err := lock.Lock()
			if err != nil {
				t.Error()
			}

			time.Sleep(lockTime)

			if err = lock.Unlock(); err != nil {
				t.Error()
			}

			wg.Done()
		}()
	}

	wg.Wait()

	// Each goroutine held the lock for lockTime, so the
	// test duration should be at least lockTime * count
	duration := time.Since(startTime)
	if int(duration/lockTime) < count {
		t.Error()
	}
}

func TestObtainLockTimeout(t *testing.T) {
	lock := FileLock{Path: "/tmp/.test2.lock", Timeout: time.Second * 1}
	if err := lock.Lock(); err != nil {
		t.Error()
	}

	defer lock.Unlock()

	lock2 := FileLock{Path: "/tmp/.test2.lock", Timeout: time.Millisecond * 10}
	err := lock2.Lock()
	if err != ErrLockTimeout {
		t.Error()
	}
}

func TestObtainLockTimeoutReleasesEventuallyObtainedLock(t *testing.T) {
	lock := FileLock{Path: "/tmp/.test3.lock", Timeout: time.Second * 1}
	if err := lock.Lock(); err != nil {
		t.Error()
	}

	lock2 := FileLock{Path: "/tmp/.test3.lock", Timeout: time.Millisecond * 10}
	if err := lock2.Lock(); err != ErrLockTimeout {
		t.Error()
	}

	// Release the first lock, this causes the
	// second lock to be obtained by the blocking goroutine,
	// and then be released straight away.
	if err := lock.Unlock(); err != nil {
		t.Error()
	}

	// Given that the first lock has been released, and
	// the second lock timed out and should have been released
	// as soon as it was obtained by the still-blocking goroutine,
	// a new lock should succeed.
	lock3 := FileLock{Path: "/tmp/.test3.lock", Timeout: time.Millisecond * 10}
	if err := lock3.Lock(); err != nil {
		t.Error()
	}
	lock3.Unlock()
}

func TestLockFilePermissionDenied(t *testing.T) {
	lock := FileLock{Path: "/.test.lock", Timeout: time.Second * 1}
	if err := lock.Lock(); err == nil {
		t.Error()
	}
}

func TestUnlockWhenNotLocked(t *testing.T) {
	lock := FileLock{Path: "/tmp/test.lock", Timeout: time.Second * 1}
	err := lock.Unlock()
	if err != ErrNotLocked {
		t.Error()
	}
}
