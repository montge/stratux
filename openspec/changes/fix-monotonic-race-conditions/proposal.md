## Why
The `monotonic` struct in `main/monotonic.go` has race conditions that prevent enabling the `-race` flag in CI tests. The `Watcher()` goroutine modifies fields (Milliseconds, Time, RealTime) while other methods read them without synchronization.

This blocks:
- Running tests with `-race` flag for better concurrency bug detection
- SonarCloud from detecting potential concurrency issues

## What Changes
- Add mutex synchronization to the `monotonic` struct
- Protect all reads and writes to shared fields
- Enable `-race` flag in CI workflow

## Impact
- Affected specs: core (thread safety)
- Affected code: `main/monotonic.go`, `.github/workflows/ci.yml`
- Risk: Low - straightforward mutex addition
- Performance: Minimal impact (10ms ticker, low contention)

## Race Condition Analysis

### Shared Fields (written by Watcher goroutine)
- `Milliseconds uint64`
- `Time time.Time`
- `RealTime time.Time`
- `realTimeSet bool`

### Unsynchronized Readers
- `Since()` - reads `Time`
- `HumanizeTime()` - reads `Time`
- `Unix()` - reads via `Since()`
- `HasRealTimeReference()` - reads `realTimeSet`
- `SetRealTimeReference()` - reads/writes `realTimeSet`, `RealTime`

### Solution
Add `sync.RWMutex` to the struct and use `RLock()`/`RUnlock()` for readers, `Lock()`/`Unlock()` for writers.
