# Go file locking

## Overview

Exclusive file locks with timeouts using `flock`.

## Example

```
  // Try to obtain an exclusive file lock with a timeout
  lock := filelock.FileLock{Path: "example.lock", Timeout: time.Second * 1}
  err := lock.Lock()
  if err != nil {
    if err == ErrLockTimeout {
      // Lock timeout
    }

    // Other error
  }

  // Lock obtained successfully, must be released.
  defer lock.Unlock()
```

See tests for more examples.

## License

Apache 2.0
