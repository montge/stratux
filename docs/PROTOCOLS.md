# Stratux Protocol Documentation

This document describes the aviation protocols that Stratux receives, parses, and transmits.

## Table of Contents

1. [NMEA-0183 (GPS)](#nmea-0183-gps)
2. [FLARM NMEA Extensions](#flarm-nmea-extensions)
3. [OGN/APRS](#ognaprs)
4. [1090 MHz ADS-B (dump1090)](#1090-mhz-ads-b-dump1090)
5. [978 MHz UAT](#978-mhz-uat)
6. [GDL90](#gdl90)
7. [X-Plane](#x-plane)

---

## NMEA-0183 (GPS)

Stratux receives GPS position data via NMEA-0183 serial protocol.

### Format

NMEA sentences are ASCII strings with the format:
```
$<talker><sentence>,<field1>,<field2>,...,<fieldN>*<checksum>\r\n
```

- **Talker ID**: 2 characters (GP=GPS, GN=GNSS, GL=GLONASS)
- **Sentence ID**: 3 characters identifying the sentence type
- **Fields**: Comma-separated data fields
- **Checksum**: XOR of all characters between $ and *

### Supported Sentences

#### GPRMC - Recommended Minimum Navigation Information

```
$GPRMC,<time>,<status>,<lat>,<N/S>,<lon>,<E/W>,<speed>,<track>,<date>,<magvar>,<E/W>*<checksum>
```

**Example:**
```
$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A
```

**Fields:**
1. UTC time (HHMMSS)
2. Status (A=valid, V=invalid)
3. Latitude (DDMM.MMMM)
4. N/S indicator
5. Longitude (DDDMM.MMMM)
6. E/W indicator
7. Speed over ground (knots)
8. Track angle (degrees true)
9. Date (DDMMYY)
10. Magnetic variation (degrees)
11. E/W indicator for mag var

#### GPGGA - Global Positioning System Fix Data

```
$GPGGA,<time>,<lat>,<N/S>,<lon>,<E/W>,<quality>,<numSV>,<HDOP>,<alt>,M,<sep>,M,<diffAge>,<diffStation>*<checksum>
```

**Example:**
```
$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47
```

**Fields:**
1. UTC time (HHMMSS)
2. Latitude (DDMM.MMMM)
3. N/S indicator
4. Longitude (DDDMM.MMMM)
5. E/W indicator
6. GPS quality (0=invalid, 1=GPS, 2=DGPS)
7. Number of satellites
8. HDOP (horizontal dilution of precision)
9. Altitude above MSL (meters)
10. Altitude units (M)
11. Geoid separation (meters)
12. Separation units (M)
13. Age of differential GPS data
14. Differential reference station ID

#### GPGSA - GPS DOP and Active Satellites

```
$GPGSA,<mode1>,<mode2>,<sv1>,<sv2>,...,<sv12>,<PDOP>,<HDOP>,<VDOP>*<checksum>
```

**Example:**
```
$GPGSA,A,3,01,02,03,04,05,06,07,08,,,,,2.0,0.9,1.8*30
```

**Fields:**
1. Mode (A=auto, M=manual)
2. Fix type (1=none, 2=2D, 3=3D)
3-14. PRN numbers of satellites used
15. PDOP (position dilution of precision)
16. HDOP (horizontal dilution of precision)
17. VDOP (vertical dilution of precision)

#### GPGSV - GPS Satellites in View

```
$GPGSV,<numMsg>,<msgNum>,<numSV>,<sv1>,<elev1>,<az1>,<snr1>,...*<checksum>
```

**Example:**
```
$GPGSV,3,1,12,01,85,045,45,02,65,135,42,03,55,225,40,04,45,315,38*75
```

**Fields:**
1. Total number of GSV messages
2. Message number
3. Total number of satellites in view
4. Satellite PRN number
5. Elevation (0-90 degrees)
6. Azimuth (0-359 degrees)
7. SNR (signal-to-noise ratio, 0-99 dB)
(Repeats for up to 4 satellites per message)

### Implementation

**Source:** `main/gps.go`

**Key Functions:**
- `processNMEALine()` - Main NMEA parser dispatcher
- `processNMEAGGA()` - Parses GPGGA sentences
- `processNMEARMC()` - Parses GPRMC sentences
- `processNMEAGSA()` - Parses GPGSA sentences
- `processNMEAGSV()` - Parses GPGSV sentences

---

## FLARM NMEA Extensions

FLARM adds proprietary NMEA sentences for collision avoidance data.

### PFLAU - Receive Status

**Syntax:**
```
$PFLAU,<RX>,<TX>,<GPS>,<Power>,<AlarmLevel>,<RelativeBearing>,<AlarmType>,<RelativeVertical>,<RelativeDistance>,<ID>*<checksum>
```

**Example:**
```
$PFLAU,2,1,2,1,2,180,2,-100,500,DD4711*3D
```

**Fields:**
1. **RX** - Number of received FLARM devices
2. **TX** - Transmission status (1=on, 0=off)
3. **GPS** - GPS status (0=none, 1=dead reckoning, 2=2D fix, 3=3D fix)
4. **Power** - Power status (1=on, 0=low)
5. **AlarmLevel** - Collision alarm level (0-3)
   - 0 = No alarm
   - 1 = Alarm, 13-18 seconds to impact
   - 2 = Alarm, 9-12 seconds to impact
   - 3 = Alarm, 0-8 seconds to impact
6. **RelativeBearing** - Relative bearing (degrees, 0-359)
7. **AlarmType** - Type of alarm (0-3)
8. **RelativeVertical** - Relative vertical separation (meters, positive=above)
9. **RelativeDistance** - Horizontal distance (meters)
10. **ID** - Hex ID of alarm target (6 characters)

### PFLAA - Traffic Data

**Syntax:**
```
$PFLAA,<AlarmLevel>,<RelativeNorth>,<RelativeEast>,<RelativeVertical>,<IDType>,<ID>,<Track>,<TurnRate>,<GroundSpeed>,<ClimbRate>,<AcftType>*<checksum>
```

**Example:**
```
$PFLAA,0,-10687,-22561,-10283,1,A4F2EE,136,0,269,0.0,0*4E
```

**Fields:**
1. **AlarmLevel** - Collision alarm level (0-3, same as PFLAU)
2. **RelativeNorth** - North position relative to own aircraft (meters)
3. **RelativeEast** - East position relative to own aircraft (meters)
4. **RelativeVertical** - Vertical separation (meters, positive=above)
5. **IDType** - ID type
   - 1 = ICAO 24-bit address
   - 2 = FLARM ID
   - 3 = OGN tracker
6. **ID** - Hex ID (6 characters)
7. **Track** - Ground track (degrees, 0-359)
8. **TurnRate** - Rate of turn (degrees/second)
9. **GroundSpeed** - Ground speed (knots)
10. **ClimbRate** - Climb rate (m/s, positive=climbing)
11. **AcftType** - Aircraft type code

### Aircraft Type Codes

The following codes are used in the AcftType field:

```
0 = Unknown
1 = Glider/Motor glider
2 = Tow/tug plane
3 = Helicopter/Rotorcraft
4 = Parachute
5 = Drop plane
6 = Hang glider
7 = Paraglider
8 = Powered aircraft
9 = Jet aircraft
A = UFO
B = Balloon
C = Airship
D = UAV
E = Reserved
F = Static object
```

### Implementation

**Source:** `main/flarm-nmea.go`

**Key Functions:**
- `makeFlarmPFLAUString()` - Generates PFLAU messages
- `makeFlarmPFLAAString()` - Generates PFLAA messages
- `parseFlarmPFLAU()` - Parses received PFLAU
- `parseFlarmPFLAA()` - Parses received PFLAA
- `calcAlarmLevel()` - Determines collision alarm level

**Tests:** `main/flarm-nmea_test.go`

---

## OGN/APRS

Open Glider Network uses APRS (Automatic Packet Reporting System) protocol over TCP connection to aprs.glidernet.org.

### APRS Message Format

OGN uses a specific APRS format for aircraft position beacons:

```
<protocol><id>><path>,qAS,<station>:/<time>h<latitude>/<longitude>/<track>/<speed>/A=<altitude> !W<precision>! id<details>
```

### Regex Pattern

```regex
(?P<protocol>ICA|FLR|SKY|PAW|OGN|RND|FMT|MTK|XCG|FAN|FNT)(?P<id>[\dA-Z]{6})>
[A-Z]+,qAS,([\d\w]+):/
(?P<time>\d{6})h(?P<longitude>\d*\.?\d*[NS])[/\\](?P<lattitude>\d*\.?\d*[EW])
\D
((?P<track>\d{3})/(?P<speed>\d{3})/A=(?P<altitude>\d*))*
(\s!W(?P<lonlatprecision>\d+)!\s)*
(id(?P<id>[\dA-F]{8}))*
```

### Example Message

```
ICA3E7868>OGNTRK,qAS,LSZF:h/122456h4659.16N/00653.77E'/229/096/A=001500 !W35! id3E3E7868 +198fpm +0.0rot
```

**Breakdown:**
- **ICA3E7868** - Protocol (ICA) + ICAO address (3E7868)
- **OGNTRK** - Destination
- **qAS,LSZF** - Quality (qAS = from aircraft via station), Station ID
- **/122456h** - Time (HHMMSS UTC)
- **4659.16N/00653.77E** - Position (DDMM.MM N/S, DDDMM.MM E/W)
- **229/096** - Track (degrees) / Speed (knots)
- **A=001500** - Altitude (feet)
- **!W35!** - Position precision
- **id3E7868** - Details field (hex encoded)

### Details Field Format

The details field is 8 hex characters encoding:

**First 2 hex chars (1 byte):**
- Bits 0-1: Address type (0=Random, 1=ICAO, 2=FLARM, 3=OGN)
- Bits 2-5: Aircraft type (0-15, similar to FLARM codes)
- Bits 6-7: Reserved

**Remaining 6 hex chars:** Tracking info

### Protocol Prefixes

- **ICA** - ICAO 24-bit address
- **FLR** - FLARM ID
- **OGN** - OGN tracker ID
- **SKY** - SkyTraxx
- **PAW** - PilotAware
- **RND** - Random/anonymous ID
- **FMT** - Fanet
- **MTK** - MTK
- **XCG** - XCG
- **FAN** - Fanet+
- **FNT** - Fanet tracker

### Implementation

**Source:** `main/ogn-aprs.go`

**Key Functions:**
- `aprsListen()` - Connects to APRS server and receives beacons
- `parseAprsMessage()` - Parses APRS message with regex
- `importOgnTrafficMessage()` - Converts to internal traffic format

---

## 1090 MHz ADS-B (dump1090)

Stratux uses dump1090 to receive 1090 MHz ADS-B (Mode-S Extended Squitter) messages.

### JSON Format

dump1090 outputs aircraft data as JSON objects. See `dump1090/README-json.md` for complete specification.

**Example:**
```json
{
  "hex": "A12345",
  "flight": "UAL123  ",
  "alt_baro": 35000,
  "alt_geom": 35200,
  "gs": 450,
  "ias": 280,
  "tas": 450,
  "track": 270,
  "baro_rate": 0,
  "lat": 47.4502,
  "lon": -122.3088,
  "nic": 8,
  "rc": 186,
  "seen_pos": 0.1,
  "version": 2,
  "nac_p": 10,
  "nac_v": 2,
  "sil": 3,
  "sil_type": "perhour",
  "messages": 150,
  "seen": 0.1,
  "rssi": -25.5
}
```

**Key Fields:**
- **hex** - ICAO 24-bit address (6 hex digits)
- **flight** - Callsign (8 chars, space-padded)
- **alt_baro** - Barometric altitude (feet)
- **alt_geom** - GNSS altitude (feet)
- **gs** - Ground speed (knots)
- **track** - True track (degrees 0-359)
- **lat, lon** - Position (decimal degrees)
- **nic** - Navigation Integrity Category
- **rssi** - Signal strength (dBFS)

### Standards References

- **DO-260B** - MOPS for 1090 MHz ADS-B
- Section references in JSON format refer to DO-260B sections

### Implementation

**Source:** `main/traffic.go`

**Key Functions:**
- `parseDump1090Message()` - Parses dump1090 JSON output
- `esListen()` - Connects to dump1090 on port 30006

**Tests:** `main/traffic_test.go`

---

## 978 MHz UAT

Universal Access Transceiver (UAT) is used for ADS-B in the US on 978 MHz.

### Format

UAT messages are decoded by the dump978 library (C library with Go bindings).

**Frame Types:**
- **Basic UAT** - Position, velocity, altitude
- **Long UAT** - Includes FIS-B weather data

### Implementation

**Source:** 
- `main/uatparse.go` - Go UAT parser
- `dump978/` - C library for UAT decoding

**Key Functions:**
- `parseUATMessage()` - Main UAT parser
- `handleUatMessage()` - Processes decoded UAT frames

**Standards References:**
- **DO-282B** - MOPS for UAT ADS-B

**Tests:** `uatparse/uatparse_test.go`

---

## GDL90

GDL90 is the output protocol Stratux uses to send traffic and weather to EFB applications.

### Message Format

Binary protocol with the following structure:
```
<FLAG><MessageID><Data...><CRC><FLAG>
```

- **FLAG** - 0x7E (frame delimiter)
- **MessageID** - 1 byte identifying message type
- **Data** - Variable length payload
- **CRC** - 2-byte CRC
- **Byte stuffing** - 0x7D escape character for data containing 0x7E or 0x7D

### Message Types

```
0x00 - Heartbeat
0x01 - Initialization
0x07 - Uplink Data
0x0A - Ownship Report
0x0B - Ownship Geometric Altitude
0x14 - Traffic Report
0x15 - GPS Time (optional)
0x65 - AHRS (Foreflight extension)
```

### Heartbeat Message (0x00)

```
7E 00 <status1> <status2> <timestamp> <message_counts> <CRC> 7E
```

**Status Byte 1:**
- Bit 0: GPS valid
- Bit 1: WAAS available
- Bit 5: UAT initialized
- Bit 7: Reserved

**Status Byte 2:**
- Bit 0: Timestamp valid
- Bit 4: AHRS valid
- Bit 5: UAT time valid

### Traffic Report (0x14)

```
7E 14 <status> <type> <address> <lat> <lon> <alt> <misc> <nav_integrity> <nav_accuracy> <hvel> <vvel> <track> <emitter> <callsign> <code> <CRC> 7E
```

**Fields:**
- **status** - Alert status and address type
- **type** - Traffic type (ADS-B, TIS-B, etc.)
- **address** - ICAO address (3 bytes)
- **lat/lon** - Position (scaled integers)
- **alt** - Altitude (pressure or geometric)
- **hvel** - Horizontal velocity (knots)
- **vvel** - Vertical velocity (ft/min, scaled)
- **track** - Ground track (degrees)
- **callsign** - 8 ASCII characters

### Implementation

**Source:** `main/gen_gdl90.go`

**Key Functions:**
- `makeHeartbeat()` - Creates heartbeat messages
- `makeTrafficReport()` - Creates traffic messages
- `makeOwnshipReport()` - Creates ownship messages
- `makeAHRSGDL90Report()` - Creates AHRS messages (ForeFlight extension)

**Standards References:**
- GDL90 specification (Garmin proprietary, widely adopted)

**Tests:** `main/gen_gdl90_test.go`

---

## X-Plane

Stratux can output data in X-Plane format for flight simulator integration.

### Message Types

#### XGPS - GPS Position

```
XGPSStratux,<lon>,<lat>,<alt_m>,<track>,<speed_mps>
```

**Fields:**
- **lon** - Longitude (decimal degrees)
- **lat** - Latitude (decimal degrees)
- **alt_m** - Altitude MSL (meters)
- **track** - Track (degrees, 0-359.9999)
- **speed_mps** - Speed (meters/second)

#### XATT - Attitude

```
XATTStratux,<heading>,<pitch>,<roll>,<hx>,<hy>,<hz>,<ax>,<ay>,<az>,<vx>,<vy>,<vz>
```

**Fields:**
- **heading** - Heading (degrees)
- **pitch** - Pitch (degrees, positive=nose up)
- **roll** - Roll (degrees, positive=right wing down)
- **hx/hy/hz** - Magnetic heading vector
- **ax/ay/az** - Acceleration (G)
- **vx/vy/vz** - Angular velocity (degrees/s)

#### XTRAFFIC - Traffic

```
XTRAFFICStratux,<lat>,<lon>,<alt_m>,<vvel_mps>,<airborne>,<track>,<speed_mps>,<hdist_m>,<bearing>,<callsign>
```

### Implementation

**Source:** `main/gen_gdl90.go`

**Key Functions:**
- `createXPlaneGpsMsg()` - GPS message
- `createXPlaneAttitudeMsg()` - Attitude message
- `createXPlaneTrafficMsg()` - Traffic message
- `convertKnotsToXPlaneSpeed()` - Unit conversion

**Tests:** `main/xplane_test.go`

---

## References

### External Standards

- **DO-260B** - Minimum Operational Performance Standards (MOPS) for 1090 MHz ADS-B
- **DO-282B** - MOPS for Universal Access Transceiver (UAT) ADS-B
- **DO-278A** - Software Integrity Assurance for Ground Systems
- **NMEA-0183** - Standard for marine electronics communication

### Stratux Documentation

- `dump1090/README-json.md` - Complete dump1090 JSON format
- `docs/TEST_PROCEDURES.md` - DO-278A SAL-3 testing procedures
- `docs/app-vendor-integration.md` - API documentation for EFB apps
- `main/testdata/README.md` - Integration testing with trace files

### Source Code

All protocol parsers are in the `main/` package:
- `gps.go` - NMEA parser
- `flarm-nmea.go` - FLARM parser and generator
- `ogn-aprs.go` - OGN/APRS parser
- `traffic.go` - dump1090 JSON parser
- `uatparse.go` - UAT parser
- `gen_gdl90.go` - GDL90 generator
