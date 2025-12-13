## 1. Implementation
- [ ] 1.1 Add sync.RWMutex field to monotonic struct
- [ ] 1.2 Add Lock/Unlock to Watcher() for writes
- [ ] 1.3 Add RLock/RUnlock to Since() for reads
- [ ] 1.4 Add RLock/RUnlock to HumanizeTime() for reads
- [ ] 1.5 Add RLock/RUnlock to Unix() for reads
- [ ] 1.6 Add RLock/RUnlock to HasRealTimeReference() for reads
- [ ] 1.7 Add Lock/Unlock to SetRealTimeReference() for writes

## 2. CI Update
- [ ] 2.1 Enable -race flag in CI workflow test command
- [ ] 2.2 Verify tests pass with -race flag

## 3. Verification
- [ ] 3.1 Run tests locally with -race flag
- [ ] 3.2 Verify no race conditions detected
- [ ] 3.3 Benchmark to ensure minimal performance impact
