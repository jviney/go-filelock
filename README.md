# Go file locking

## Overview

Exclusive file locks with timeouts.

## Example

```
  // Try to obtain an exclusive lock with a short timeout
  lock := filelock.Obtain("example.lock", time.Second * 1)

  // Lock obtained successfully.
  // Must be released.
  if lock.IsLocked() {
    // code here ...
    lock.Release()
  }

  // Unable to get the lock before the specified timeout.
  if lock.State == filelock.LockTimeout {
    log.Printf("lock timeout")
  }

  // An error occurred
  if lock.State == filelock.LockError {
    log.Printf(lock.Error.Error())
  }
```

See tests for more examples.

## License

Apache 2.0
