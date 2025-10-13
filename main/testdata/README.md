# Test Data Fixtures

This directory contains test data fixtures for Stratux unit tests.

## Organization

- `adsb/` - Sample ADS-B messages (1090 MHz)
- `uat/` - Sample UAT messages (978 MHz)
- `ogn/` - Sample OGN/FLARM messages (868 MHz)
- `gps/` - Sample GPS NMEA sentences
- `gdl90/` - Expected GDL90 output messages
- `config/` - Sample configuration files

## File Formats

- `.bin` - Binary message data
- `.hex` - Hexadecimal message dumps
- `.txt` - Text-based data (NMEA, etc.)
- `.json` - JSON configuration or expected results

## Usage

Test files should reference fixtures using relative paths:

```go
data, err := os.ReadFile("testdata/adsb/sample_traffic.bin")
```

## Conventions

- Filename format: `{description}_{variant}.{ext}`
- Examples:
  - `adsb_valid_position.hex`
  - `adsb_invalid_crc.hex`
  - `gps_fix3d.txt`
  - `gdl90_heartbeat_expected.bin`
