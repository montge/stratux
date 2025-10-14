# Stratux Test Data

This directory contains sample trace files and test data for integration testing.

## Trace Files

Trace files are gzipped CSV files that record real or simulated Stratux input data. They can be used for:
- **Integration testing** - Replay recorded scenarios without hardware
- **Debugging** - Reproduce specific issues
- **Performance testing** - Test with known data sets

### Trace File Format

Each trace file is a gzipped CSV with 3 columns:
1. **Timestamp** (RFC3339Nano format) - When the message was received
2. **Context** - Message source type (see below)
3. **Data** - The raw message content

### Supported Contexts

- `ais` - AIS (marine traffic) messages
- `nmea` - GPS NMEA-0183 sentences
- `aprs` - APRS messages
- `ogn-rx` - Open Glider Network messages
- `dump1090` - 1090 MHz ADS-B messages (JSON format)
- `godump978` - 978 MHz UAT messages
- `lowpower_uat` - Low-power UAT messages

## Directory Structure

```
testdata/
├── adsb/           - 1090 MHz ADS-B test data
│   ├── basic_adsb.trace.gz     - Simple two-aircraft scenario
│   └── generate_trace.go       - Generator script
├── uat/            - 978 MHz UAT test data
├── gps/            - GPS/GNSS test data
│   ├── basic_gps.trace.gz      - Simple GPS NMEA sequence
│   └── generate_trace.go       - Generator script
├── gdl90/          - GDL90 protocol test data
└── ogn/            - OGN/APRS test data
```

## Using Trace Files

### Manual Replay with Stratux

```bash
# Replay at normal speed
./stratuxrun -trace testdata/adsb/basic_adsb.trace.gz -traceSpeed 1.0

# Replay at 2x speed
./stratuxrun -trace testdata/adsb/basic_adsb.trace.gz -traceSpeed 2.0

# Replay only specific message types
./stratuxrun -trace testdata/adsb/basic_adsb.trace.gz -traceFilter "dump1090,nmea"
```

## Recording Trace Files

To record live data from hardware:

1. Enable trace logging in the web interface (Settings → TraceLog)
2. Fly or run Stratux with hardware connected
3. Trace files are saved to `/var/log/stratux/` as `YYYY-MM-DDTHH:MM:SS_trace.txt.gz`
4. Copy the file to the appropriate testdata subdirectory

## Sample Scenarios

### basic_adsb.trace.gz
- **Duration**: 3 seconds
- **Aircraft**: 2 (UAL123 at 35000ft, N172SP at 5500ft)
- **Use case**: Basic traffic parsing and tracking

### basic_gps.trace.gz
- **Duration**: 3 seconds
- **Messages**: GPRMC, GPGGA, GPGSA, GPGSV
- **Use case**: GPS position parsing
