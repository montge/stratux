# Additional Test Coverage Opportunities

**Question**: Are there other parts of Stratux (like the UI) that we can increase coverage on before we do refactoring?

**Answer**: Yes! There are several components beyond the main Go package that could benefit from test coverage without requiring refactoring. However, the impact and ease of testing varies significantly.

---

## Component Analysis

### 1. Web UI (JavaScript) ⭐ HIGH IMPACT

**Location**: `web/js/`, `web/plates/js/`
**Total Lines**: ~4,225 lines of JavaScript (excluding libraries)
**Current Tests**: None
**Language**: JavaScript (AngularJS 1.x framework)

#### Files Breakdown:
```
web/js/main.js              ~10,000 lines (includes minified libraries)
web/plates/js/status.js     ~800 lines
web/plates/js/traffic.js    ~500 lines
web/plates/js/gps.js        ~400 lines
web/plates/js/settings.js   ~600 lines
web/plates/js/map.js        ~600 lines
web/plates/js/radar.js      ~300 lines
web/plates/js/weather.js    ~260 lines
web/plates/js/logs.js       ~200 lines
web/plates/js/towers.js     ~73 lines
web/plates/js/developer.js  ~172 lines
web/plates/js/ahrs.js       ~268 lines
```

#### Testing Approach:

**Option A: Jest + Jasmine (Modern JavaScript testing)**
- **Pros**: Industry standard, great tooling, mocking support
- **Cons**: Requires build setup (package.json, npm, etc.)
- **Estimated setup**: 1-2 days
- **Example test**:
  ```javascript
  describe('craftService', () => {
    let service;

    beforeEach(() => {
      service = craftService();
    });

    test('getAircraftCategory returns correct category', () => {
      const aircraft = { Emitter_category: 1 };
      expect(service.getCategory(aircraft)).toBe('Light');
    });

    test('isTrafficAged correctly identifies aged aircraft', () => {
      const aircraft = { Age: 65, TargetType: 1 };
      expect(service.isTrafficAged(aircraft)).toBe(true);
    });
  });
  ```

**Option B: Selenium/Puppeteer (E2E UI testing)**
- **Pros**: Tests actual user workflows
- **Cons**: Slower, requires running Stratux server
- **Estimated setup**: 2-3 days

#### What Could Be Tested:
- ✅ **Pure JavaScript functions** (80% testable):
  - `craftService` - aircraft/vessel categorization logic
  - `getTrafficSourceColor()` - color mapping
  - `isTrafficAged()` - age calculation
  - URL builders, formatters, converters

- ⚠️ **AngularJS controllers** (50% testable):
  - Data transformation logic
  - State management
  - Requires AngularJS mocking

- ❌ **WebSocket/HTTP interactions** (harder):
  - Requires mocking `$http`, WebSocket connections
  - Integration tests better suited

**Coverage Potential**: 60-70% with unit tests + 20-30% with E2E tests

**Impact**: HIGH - The web UI is user-facing and has complex logic that could benefit from tests

**Recommendation**: ⭐ **YES - Start with Jest for unit tests** (pure functions first)

---

### 2. fancontrol Package 🔧 MEDIUM IMPACT

**Location**: `fancontrol_main/fancontrol.go`
**Total Lines**: 326 lines
**Current Tests**: None
**Language**: Go

#### What It Does:
- PID controller for CPU fan speed based on temperature
- PWM (Pulse Width Modulation) control
- Prometheus metrics export
- HTTP status endpoint

#### Testing Analysis:

**Testable Without Hardware** (30% of code):
- ✅ `fmap()` - Range mapping function (pure math)
- ✅ `readSettings()` - JSON config parsing
- ✅ `handleStatusRequest()` - HTTP handler
- ✅ Struct marshaling/unmarshaling

**Hardware-Dependent** (70% of code):
- ❌ `fanControl()` - Main loop with GPIO access
- ❌ `turnOnFanTest()` - Physical fan interaction
- ❌ `updateStats()` - Prometheus metrics goroutine
- ❌ PWM frequency/duty cycle (requires rpio library + hardware)

**Example Testable Function**:
```go
func TestFmapRangeMapping(t *testing.T) {
    tests := []struct {
        x, inMin, inMax, outMin, outMax, expected float64
    }{
        {5, 0, 10, 0, 100, 50},
        {0, 0, 10, 0, 100, 0},
        {10, 0, 10, 0, 100, 100},
        {25, 0, 100, 32, 212, 77}, // Celsius to Fahrenheit
    }

    for _, tt := range tests {
        result := fmap(tt.x, tt.inMin, tt.inMax, tt.outMin, tt.outMax)
        if math.Abs(result-tt.expected) > 0.001 {
            t.Errorf("fmap(%f, %f, %f, %f, %f) = %f, want %f",
                tt.x, tt.inMin, tt.inMax, tt.outMin, tt.outMax, result, tt.expected)
        }
    }
}
```

**Coverage Potential**: 30-40% (without hardware), 80%+ (with mocking/refactoring)

**Impact**: MEDIUM - Critical for hardware health but simple logic

**Recommendation**: ⚠️ **MAYBE - Low-hanging fruit only**
- Test `fmap()`, `readSettings()`, config handling
- Don't try to test hardware interaction
- Better to add hardware integration tests later

---

### 3. sensors Package 🌡️ LOW IMPACT

**Location**: `sensors/`
**Total Lines**: ~1,123 lines
**Current Tests**: None
**Language**: Go

#### What It Contains:
- `imu.go` - IMU interface (17 lines) ✅ Testable
- `pressure.go` - Pressure sensor interface (10 lines) ✅ Testable
- `bmp280.go` - Bosch BMP280 pressure sensor driver (81 lines) ❌ Hardware
- `bmp388.go` - Bosch BMP388 pressure sensor driver (79 lines) ❌ Hardware
- `bmp388/bmp388.go` - BMP388 low-level driver (217 lines) ❌ Hardware
- `bmp388/registers.go` - Register definitions (85 lines) ✅ Testable
- `mpu9250.go` - MPU9250 IMU driver (98 lines) ❌ Hardware
- `icm20948.go` - ICM20948 IMU driver (98 lines) ❌ Hardware

#### Testing Analysis:

**Testable** (20%):
- ✅ Interface definitions (`IMUReader`, `PressureReader`)
- ✅ Register constants and bit masks
- ✅ Coordinate transformation math (if any)

**Not Testable Without Hardware** (80%):
- ❌ I2C communication
- ❌ Sensor initialization sequences
- ❌ Data reading from physical sensors
- ❌ Calibration procedures

**Coverage Potential**: 20-30% (constants/interfaces only)

**Impact**: LOW - Mostly hardware drivers, small codebase

**Recommendation**: ❌ **NO - Skip for now**
- Very hardware-dependent
- Small codebase doesn't justify effort
- Better to do integration testing with actual hardware
- Would need significant refactoring to test properly

---

### 4. godump978 Package 📡 LOW IMPACT

**Location**: `godump978/`
**Total Lines**: ~60 lines
**Current Tests**: None
**Language**: Go (CGO wrapper)

#### What It Does:
- CGO wrapper for `libdump978.so` (C library)
- Demodulates 978MHz UAT signals
- Provides Go interface to C demodulator

#### Testing Analysis:

**Structure**:
```go
// godump978.go (62 lines)
func ProcessData(buf []byte)              // Calls C function
func ProcessDataFromChannel()             // Goroutine wrapper
var InChan = make(chan []byte, 100)       // Channel

// godump978_exports.go (44 lines)
//export dump978Cb                         // CGO callback
```

**Testable** (20%):
- ✅ Channel operations (can send/receive on `InChan`)
- ✅ Package version constant

**Not Testable Without C Library** (80%):
- ❌ `ProcessData()` - Calls C `process_data()`
- ❌ `dump978Cb()` - C callback from demodulator
- ❌ Requires libdump978.so and IQ samples

**Coverage Potential**: 10-20% (very thin wrapper)

**Impact**: LOW - Thin wrapper, already integration tested via main

**Recommendation**: ❌ **NO - Skip**
- Too thin to justify separate tests
- Already tested via main package integration tests
- C library has its own test suite

---

## Summary: Where to Focus

### Priority 1: Web UI JavaScript ⭐⭐⭐
**Effort**: Medium (2-3 days setup + testing)
**Impact**: HIGH (4,000+ lines of testable logic)
**Coverage Gain**: 60-80% of web UI code
**Benefits**:
- User-facing code quality improvement
- Catch UI bugs before deployment
- Enable safe refactoring of UI logic
- Modern CI/CD integration

**Recommended Approach**:
1. Set up Jest + package.json (1 day)
2. Test pure JavaScript functions first (1 day)
   - `craftService` methods
   - Color mapping functions
   - Age calculation logic
3. Test AngularJS controllers with mocks (1-2 days)
4. Add E2E tests with Puppeteer (optional, 2-3 days)

**Files to Create**:
```
web/package.json              # Node.js dependencies
web/jest.config.js            # Jest configuration
web/.babelrc                  # Babel for ES6 support
web/tests/
├── craftService.test.js      # Service tests
├── statusCtrl.test.js        # Controller tests
└── trafficCtrl.test.js       # Controller tests
.github/workflows/
└── web-tests.yml             # CI for web tests
```

---

### Priority 2: fancontrol Utility Functions ⭐
**Effort**: Low (0.5-1 day)
**Impact**: MEDIUM (300 lines, critical hardware component)
**Coverage Gain**: 30-40% of fancontrol code
**Benefits**:
- Quick wins with pure functions
- Validate PID controller math
- Test config handling

**Recommended Approach**:
1. Create `fancontrol_main/fancontrol_test.go`
2. Test `fmap()` with comprehensive cases
3. Test `readSettings()` with mock JSON files
4. Test struct marshaling/unmarshaling
5. Stop there (don't try to test hardware)

**Example Tests**:
```go
// fancontrol_main/fancontrol_test.go
func TestFmap(t *testing.T) { ... }
func TestReadSettings(t *testing.T) { ... }
func TestFanControlStructMarshaling(t *testing.T) { ... }
```

---

### Priority 3: sensors (Skip) ❌
**Effort**: High (would require refactoring)
**Impact**: LOW (1,123 lines, mostly hardware drivers)
**Coverage Gain**: 20-30%
**Recommendation**: Skip - not worth it without refactoring

---

### Priority 4: godump978 (Skip) ❌
**Effort**: Medium (requires C library mocking)
**Impact**: LOW (60 lines, thin wrapper)
**Coverage Gain**: 10-20%
**Recommendation**: Skip - already tested via integration

---

## Overall Coverage Impact Estimate

| Component | Current | With Tests | Effort | Priority |
|-----------|---------|------------|--------|----------|
| **main** | 9.4% | →40-50% (roadmap) | High | ✅ In Progress |
| **common** | 90.2% | ✅ Complete | - | ✅ Done |
| **uatparse** | 29.7% | →80% (roadmap) | Medium | ⏳ Planned |
| **Web UI (JS)** | 0% | →60-80% | Medium | ⭐ Recommend |
| **fancontrol** | 0% | →30-40% | Low | ⭐ Quick Win |
| **sensors** | 0% | →20-30% | High | ❌ Skip |
| **godump978** | 0% | →10-20% | Medium | ❌ Skip |

---

## Recommended Action Plan

### Phase A: Web UI Testing (High Value) ⭐
**Duration**: 4-6 days
**Impact**: Major improvement in user-facing code quality

1. **Day 1**: Set up Jest testing framework
   - Create `package.json` with Jest, Babel
   - Configure `jest.config.js`
   - Add npm scripts for testing
   - Update GitHub Actions CI

2. **Day 2-3**: Unit tests for pure functions
   - Test `craftService` methods (20+ tests)
   - Test color mapping functions
   - Test age calculation logic
   - Target: 200-300 lines of test code

3. **Day 4-5**: Controller tests
   - Mock AngularJS dependencies
   - Test `StatusCtrl`, `TrafficCtrl`
   - Test data transformation
   - Target: 300-400 lines of test code

4. **Day 6** (optional): E2E tests
   - Puppeteer setup
   - Test critical user flows
   - Login, settings, traffic view

**Expected Result**: 60-80% coverage of web UI, ~500-700 lines of test code

---

### Phase B: fancontrol Quick Wins (Low Effort) ⭐
**Duration**: 0.5-1 day
**Impact**: Validate critical hardware control logic

1. Create `fancontrol_main/fancontrol_test.go`
2. Test `fmap()` with 10+ test cases
3. Test `readSettings()` with mock configs
4. Test JSON marshaling

**Expected Result**: 30-40% coverage of fancontrol, ~150 lines of test code

---

## Integration with Existing Roadmap

The coverage roadmap already targets the main Go packages. Adding web UI and fancontrol tests would:

**Before**:
```
Phase 1-5: main/uatparse/common Go packages
Coverage: 9.4% → 65% (Go code only)
```

**After** (with Web UI + fancontrol):
```
Phase 1-5: Go packages (65%)
Phase 6: Web UI JavaScript (60-80%)
Phase 7: fancontrol utilities (30-40%)

Overall Project Coverage: ~70%
```

---

## Decision Matrix

### Should we test the Web UI?

**✅ YES, if**:
- User-facing bugs are a concern
- You want to enable safe UI refactoring
- You have 4-6 days available
- CI/CD for frontend is valuable

**❌ NO, if**:
- Focus must stay on Go/backend code
- No frontend developers available
- Timeline is very tight

**My Recommendation**: ⭐ **YES - Web UI testing is high value**
- 4,000+ lines of untested user-facing code
- Complex logic that could have bugs (age calculation, categorization, colors)
- Modern testing setup would future-proof the UI

### Should we test fancontrol?

**✅ YES, if**:
- You want quick coverage wins (0.5-1 day)
- Hardware reliability is important
- PID controller math needs validation

**❌ NO, if**:
- Focus must be 100% on main package
- Hardware testing is sufficient

**My Recommendation**: ⭐ **YES - Quick win, low effort**
- Only 0.5-1 day of work
- Tests pure math functions (no hardware needed)
- Validates critical temperature control logic

### Should we test sensors or godump978?

**❌ NO**:
- Too hardware-dependent
- Low coverage potential without refactoring
- Better to focus on higher-value targets

---

## Conclusion

**Answer to original question**: YES, there are significant opportunities for test coverage improvements beyond the main Go package, specifically:

1. **Web UI (JavaScript)** - 4,000+ lines, HIGH IMPACT, MEDIUM EFFORT
2. **fancontrol utilities** - 300 lines, MEDIUM IMPACT, LOW EFFORT

Both can be tested **without refactoring** the existing code. The Web UI offers the highest value for improving overall project test coverage and quality.

**Recommended next steps**:
1. Complete current roadmap Phase 2 (Protocol Parser Integration Tests)
2. Add Web UI testing (Phase 6)
3. Add fancontrol utility tests (Phase 7)
4. Final coverage: ~70% across all components
