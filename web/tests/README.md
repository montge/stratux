# Stratux Web UI Tests

This directory contains Jest-based unit tests for the Stratux web user interface JavaScript code.

## Test Coverage (Current)

**Total Tests**: 77 tests across 2 test files
**Status**: All tests passing ✅
**Test Files**:
- `craftService.test.js` - 41 tests
- `trafficUtils.test.js` - 36 tests

## What's Tested

### craftService.test.js (41 tests)
Tests the `craftService` from `web/js/main.js` (lines 149-328), covering:

**Traffic Source Colors (6 tests)**
- ES (1090ES ADS-B) → cornflowerblue
- UAT (978MHz) → orange
- OGN (Open Glider Network) → green
- AIS (Marine traffic) → blue
- Unknown sources → gray

**Aircraft Colors (5 tests)**
- Color mapping based on source + target type combination
- Handles ES, UAT, OGN sources with different target types
- Returns white for unknown combinations

**Aircraft Categories (7 tests)**
- Emitter category mapping (Light, Heavy, Helic, Glide, Drone, etc.)
- Handles 19 different aircraft categories from GDL90 spec
- Returns '---' for unknown categories

**Vessel Categories (7 tests)**
- AIS vessel type categorization (Cargo, Passenger, Tanker, Fishing, etc.)
- Uses first and second digit of SurfaceVehicleType
- Handles 30+ vessel types from AIS spec

**Vessel Colors (6 tests)**
- Color mapping for marine vessels based on AIS type
- Different colors for cargo, passenger, tanker, fishing, etc.

**Traffic Age Detection (10 tests)**
- Determines if aircraft/vessel is aged/stale
- Aircraft: >59 seconds = aged
- AIS vessels: >900 seconds (15 minutes) = aged
- Works with Age and AgeLastAlt fields

### trafficUtils.test.js (36 tests)
Tests utility functions from `web/plates/js/traffic.js`, covering:

**UTC Time String Formatting (8 tests)**
- Converts epoch timestamp to HH:MM:SSZ format
- Handles midnight, noon, single digits with leading zeros
- Always includes 'Z' suffix for UTC

**DMS Coordinate Formatting (12 tests)**
- Converts decimal degrees to Degrees Minutes format
- Handles positive/negative coordinates
- Pads single digits with leading zeros
- Tested with real coordinates (Seattle: 47.45°N, 122.31°W)

**Aircraft Comparison Logic (8 tests)**
- Determines if two aircraft are the same based on address and type
- Handles ICAO vs non-ICAO address types
- Type 1 = non-ICAO, all others = ICAO
- Returns undefined for mixed ICAO/non-ICAO comparisons

**Edge Cases (8 tests)**
- Epoch 0, current time
- Min/max coordinates (±180° longitude, ±90° latitude)
- Small decimal values, null types

## Running Tests

### Run all tests
```bash
cd web
npm test
```

### Run tests in watch mode (auto-rerun on file changes)
```bash
npm run test:watch
```

### Run tests with coverage report
```bash
npm run test:coverage
```

### Run tests in CI mode
```bash
npm run test:ci
```

## Test Structure

Each test file follows this pattern:

1. **Function Extraction**: Pure functions are extracted from Angular services/controllers for testing
2. **Test Suites**: Grouped by functionality using `describe()`
3. **Individual Tests**: Each test uses `test()` with descriptive names
4. **Assertions**: Uses Jest's `expect()` with matchers like `toBe()`, `toContain()`, `toMatch()`

## Adding New Tests

### 1. Create a new test file
```bash
web/tests/yourFeature.test.js
```

### 2. Extract testable functions
```javascript
// Extract pure function from Angular code
const yourFunction = (input) => {
    // Implementation copied from source file
    return output;
};
```

### 3. Write tests
```javascript
describe('yourFeature - description', () => {
    test('should do something', () => {
        expect(yourFunction(input)).toBe(expectedOutput);
    });
});
```

### 4. Run tests
```bash
npm test
```

## Dependencies

- **Jest**: Testing framework
- **@babel/core**: JavaScript compiler
- **@babel/preset-env**: Babel preset for ES6+ support
- **babel-jest**: Babel integration for Jest
- **jest-environment-jsdom**: DOM environment for browser-like testing

## Configuration

### package.json
- Defines test scripts and Jest configuration
- Excludes minified files and libraries from coverage
- Uses jsdom environment for DOM testing

### .babelrc
- Configures Babel to transpile ES6+ code
- Targets current Node.js version

## Future Work

### Phase 1 (Current) ✅ COMPLETE
- ✅ Jest setup and configuration
- ✅ Pure function tests (craftService, traffic utilities)
- ✅ 77 tests covering core logic

### Phase 2 (Planned)
- [ ] AngularJS controller tests with mocks
- [ ] Test StatusCtrl, TrafficCtrl data transformation
- [ ] Mock $http, $interval, $scope
- Target: +30-40 tests

### Phase 3 (Optional)
- [ ] E2E tests with Puppeteer
- [ ] Test critical user workflows
- [ ] Login, settings, traffic view
- Target: +10-15 tests

## Coverage Goals

**Target**: 60-80% coverage of ~4,225 lines of JavaScript code

**Current Status**:
- Pure functions: Well covered (77 tests)
- Controllers: Not yet tested
- WebSocket handlers: Not yet tested
- DOM manipulation: Not yet tested

**Achievable Coverage**:
- Pure JavaScript functions: 80-90% (current focus)
- AngularJS controllers: 50-70% (with mocks)
- WebSocket/HTTP interactions: 20-30% (harder to test)
- Overall: 60-70% (realistic target)

## Continuous Integration

Tests run automatically in GitHub Actions on every push:
- Go package tests
- Web UI tests (npm test)
- Coverage reports generated

See `.github/workflows/ci.yml` for CI configuration.

## Notes

- Tests use standalone function copies to avoid Angular dependency injection complexity
- Functions are copied from source files and tested in isolation
- This approach works well for pure functions but requires mocking for Angular-dependent code
- Future controller tests will need angular-mocks or similar library

## References

- Jest Documentation: https://jestjs.io/
- Babel Documentation: https://babeljs.io/
- Testing AngularJS: https://docs.angularjs.org/guide/unit-testing
