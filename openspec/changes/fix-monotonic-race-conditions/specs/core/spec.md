## ADDED Requirements

### Requirement: Thread-Safe Monotonic Clock
The monotonic clock SHALL be thread-safe, allowing concurrent access from multiple goroutines without data races.

#### Scenario: Concurrent reads during ticker updates
- **WHEN** the Watcher goroutine updates the time
- **AND** another goroutine calls Since() or Unix()
- **THEN** no data race SHALL occur
- **AND** the reader SHALL receive a consistent value

#### Scenario: SetRealTimeReference thread safety
- **WHEN** SetRealTimeReference is called
- **AND** the Watcher goroutine is running
- **THEN** no data race SHALL occur
- **AND** the real time reference SHALL be set atomically

### Requirement: Race Detection in CI
The CI pipeline SHALL run tests with the Go race detector enabled.

#### Scenario: Race detector enabled
- **WHEN** CI runs go test
- **THEN** the -race flag SHALL be included
- **AND** any detected races SHALL fail the build
