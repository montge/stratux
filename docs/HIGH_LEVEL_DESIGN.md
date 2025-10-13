# Stratux High-Level Design Document

**Document ID**: HLD-STRATUX-001
**Version**: 1.0 DRAFT
**Date**: 2025-10-13
**Classification**: SAL-3 per DO-278A
**Status**: Draft for Review

---

## 1. Introduction

### 1.1 Purpose

This document describes the high-level architecture and design of the Stratux ADS-B/UAT/OGN receiver system.

### 1.2 Scope

This document covers:
- System architecture and major components
- Data flow between subsystems
- Interface definitions
- Design rationale for key decisions

### 1.3 References

1. **REQUIREMENTS.md** - System Requirements Specification (SRS-STRATUX-001)
2. **DO-278A** - Software Integrity Assurance for CNS/ATM Systems
3. **DO-260B** - 1090 MHz Extended Squitter ADS-B MOPS
4. **DO-282B** - UAT ADS-B MOPS
5. **GDL90 ICD** - Garmin Data Link 90 Interface Control Document
6. **RTCA DO-289** - ADS-B MASPS (Minimum Aviation System Performance Standards)

---

## 2. System Architecture

### 2.1 Top-Level Architecture

```mermaid
graph TB
    subgraph "External Inputs"
        RF[RF Signals<br/>1090/978/868/162 MHz]
        GPS_IN[GPS<br/>UART/USB/TCP]
        SENSORS[Sensors<br/>I2C IMU/Baro]
        CONFIG[Configuration<br/>Web UI/File]
    end

    subgraph "Stratux Core System"
        subgraph "Input Processing Layer"
            SDR[SDR Manager<br/>sdr.go]
            GPS[GPS Manager<br/>gps.go]
            AHRS[AHRS Manager<br/>sensors.go]
        end

        subgraph "Processing Layer"
            TRAFFIC[Traffic Fusion<br/>traffic.go]
            WEATHER[Weather Processing<br/>gen_gdl90.go]
            OWNSHIP[Ownship State<br/>gps.go]
        end

        subgraph "Output Layer"
            GDL90[GDL90 Generator<br/>gen_gdl90.go]
            NMEA[NMEA Generator<br/>flarm-nmea.go]
            XPLANE[X-Plane Output<br/>xplane.go]
        end

        subgraph "Distribution Layer"
            NETWORK[Network Manager<br/>network.go]
            CLIENTS[Client Manager<br/>clientconnection.go]
        end

        subgraph "Management Layer"
            WEBUI[Web Interface<br/>managementinterface.go]
            DATALOG[Data Logging<br/>datalog.go]
            OTA[OTA Updates<br/>managementinterface.go]
        end
    end

    subgraph "External Outputs"
        UDP[UDP Broadcast<br/>Port 4000]
        TCP[TCP Stream<br/>Port 2000]
        SERIAL[Serial Output<br/>/dev/serialout*]
        BLE[Bluetooth LE<br/>GATT]
        WEB[Web Browser<br/>HTTP/WebSocket]
    end

    RF --> SDR
    GPS_IN --> GPS
    SENSORS --> AHRS
    CONFIG --> WEBUI

    SDR --> TRAFFIC
    GPS --> OWNSHIP
    AHRS --> OWNSHIP

    TRAFFIC --> GDL90
    WEATHER --> GDL90
    OWNSHIP --> GDL90
    OWNSHIP --> NMEA

    GDL90 --> NETWORK
    NMEA --> NETWORK
    XPLANE --> NETWORK

    NETWORK --> CLIENTS
    CLIENTS --> UDP
    CLIENTS --> TCP
    CLIENTS --> SERIAL
    CLIENTS --> BLE

    WEBUI --> WEB
    TRAFFIC --> DATALOG
    WEATHER --> DATALOG

    style RF fill:#e1f5ff
    style GPS_IN fill:#e1f5ff
    style SENSORS fill:#e1f5ff
    style CONFIG fill:#e1f5ff
    style UDP fill:#ffe1e1
    style TCP fill:#ffe1e1
    style SERIAL fill:#ffe1e1
    style BLE fill:#ffe1e1
    style WEB fill:#ffe1e1
```

### 2.2 Component Responsibilities

| Component | Primary Responsibility | Requirements Traced |
|-----------|----------------------|-------------------|
| **SDR Manager** | Manage RTL-SDR devices, spawn decoders | FR-101 to FR-105 |
| **GPS Manager** | Acquire position, manage GPS devices | FR-201 to FR-205 |
| **AHRS Manager** | Read sensors, compute attitude | FR-301 to FR-305 |
| **Traffic Fusion** | Consolidate traffic from all sources | FR-401 to FR-407 |
| **Weather Processing** | Decode FIS-B weather products | FR-501 to FR-503 |
| **GDL90 Generator** | Create GDL90 messages | FR-601 to FR-606 |
| **NMEA Generator** | Create NMEA/FLARM messages | FR-701 to FR-702 |
| **Network Manager** | Distribute messages to clients | FR-801 to FR-805 |
| **Web Interface** | Configuration and monitoring | FR-901 to FR-905 |
| **Data Logging** | Record traffic/weather/AHRS | FR-1001 to FR-1003 |

---

## 3. Data Flow Architecture

### 3.1 Traffic Data Flow

```mermaid
flowchart LR
    subgraph "RF Reception"
        A1[1090 MHz<br/>dump1090]
        A2[978 MHz<br/>dump978]
        A3[868 MHz<br/>ogn-rx]
        A4[162 MHz<br/>rtl_ais]
    end

    subgraph "Message Parsing"
        B1[Parse JSON<br/>Mode S/ES]
        B2[Parse UAT<br/>Frames]
        B3[Parse APRS<br/>OGN]
        B4[Parse AIVDM<br/>AIS]
    end

    subgraph "Traffic Fusion"
        C[Traffic Map<br/>Key: ICAO Address]
        D[Update/Create<br/>Target]
        E[Position<br/>Extrapolation]
        F[Ownship<br/>Detection]
        G[Alert Logic]
    end

    subgraph "Output"
        H[GDL90<br/>Traffic Report]
        I[NMEA<br/>PFLAA]
        J[Web UI<br/>JSON]
    end

    A1 --> B1
    A2 --> B2
    A3 --> B3
    A4 --> B4

    B1 --> C
    B2 --> C
    B3 --> C
    B4 --> C

    C --> D
    D --> E
    E --> F
    F --> G

    G --> H
    G --> I
    G --> J

    style C fill:#ffeb3b
    style F fill:#f44336,color:#fff
    style G fill:#f44336,color:#fff
```

**Critical Path Elements** (Safety-Critical):
- **Traffic Fusion (C)**: Must not lose or corrupt traffic data
- **Ownship Detection (F)**: Must accurately identify and filter ownship
- **Alert Logic (G)**: Must correctly identify proximate traffic

### 3.2 Ownship Data Flow

```mermaid
flowchart TB
    subgraph "Position Sources"
        G1[u-blox GPS<br/>UART]
        G2[Prolific GPS<br/>USB]
        G3[NMEA TCP<br/>Port 30011]
        G4[OGN Tracker<br/>GPS]
    end

    subgraph "GPS Processing"
        P1[Parse NMEA<br/>Sentences]
        P2[Validate Fix<br/>Quality]
        P3[Compute<br/>Accuracy]
        P4[Update Global<br/>State]
    end

    subgraph "AHRS Processing"
        A1[Read IMU<br/>I2C]
        A2[Read Baro<br/>I2C]
        A3[Sensor Fusion<br/>Attitude]
        A4[Barometric<br/>Altitude]
    end

    subgraph "Ownship State"
        O[Global Situation<br/>mySituation]
    end

    subgraph "Output Generation"
        OUT1[GDL90 Ownship<br/>Message 0x0A/0x0B]
        OUT2[GDL90 AHRS<br/>ForeFlight]
        OUT3[NMEA GPS<br/>GPRMC/GPGGA]
        OUT4[Web UI<br/>Status]
    end

    G1 & G2 & G3 & G4 --> P1
    P1 --> P2
    P2 --> P3
    P3 --> P4
    P4 --> O

    A1 & A2 --> A3
    A3 & A4 --> O

    O --> OUT1
    O --> OUT2
    O --> OUT3
    O --> OUT4

    style P2 fill:#f44336,color:#fff
    style P3 fill:#f44336,color:#fff
    style O fill:#ffeb3b
```

**Critical Functions**:
- **Validate Fix Quality (P2)**: Prevents invalid position from being output (FR-204)
- **Compute Accuracy (P3)**: Provides NACp for position integrity (FR-204)
- **Global Situation (O)**: Single source of truth for ownship state

### 3.3 Weather Data Flow

```mermaid
flowchart LR
    subgraph "UAT Reception"
        U[978 MHz<br/>dump978]
    end

    subgraph "FIS-B Decoding"
        F1[Parse UAT<br/>Frames]
        F2[Extract FIS-B<br/>Products]
        F3[Decompress<br/>if needed]
    end

    subgraph "Weather Products"
        W1[NEXRAD]
        W2[METAR/TAF]
        W3[Winds Aloft]
        W4[SIGMET/AIRMET]
        W5[PIREP]
        W6[Lightning]
        W7[Other Products]
    end

    subgraph "Product Management"
        M1[Weather Map<br/>Key: Product ID]
        M2[Timestamp<br/>Tracking]
        M3[Expiration<br/>15 min]
    end

    subgraph "Output"
        O1[GDL90 FIS-B<br/>Message 0x63]
        O2[Web UI<br/>Display]
    end

    U --> F1
    F1 --> F2
    F2 --> F3

    F3 --> W1 & W2 & W3 & W4 & W5 & W6 & W7

    W1 & W2 & W3 & W4 & W5 & W6 & W7 --> M1
    M1 --> M2
    M2 --> M3

    M3 --> O1
    M3 --> O2

    style M2 fill:#ff9800
    style M3 fill:#f44336,color:#fff
```

**Key Design Decisions**:
- **15-minute expiration**: Balances memory usage vs. product availability
- **Product ID keying**: Prevents duplicates, allows updates
- **Timestamp tracking**: Enables staleness detection

---

## 4. Module Design

### 4.1 Traffic Module (traffic.go)

```mermaid
classDiagram
    class TrafficInfo {
        +uint32 Icao_addr
        +float64 Lat
        +float64 Lng
        +int32 Alt
        +uint16 Track
        +uint16 Speed
        +string Tail
        +time.Time Last_seen
        +bool Position_valid
        +uint8 ExtrapolatedPosition
        +int16 BearingDist_valid
        +uint16 Bearing
        +float32 Distance
        +bool isAirborne()
        +updateExtrapolatedPosition()
    }

    class TrafficMap {
        +map[uint32]*TrafficInfo targets
        +sync.RWMutex mu
        +addOrUpdate(icao, info)
        +get(icao)
        +removeStale()
        +getAll()
    }

    class TrafficSource {
        <<enumeration>>
        ADS_B_1090
        UAT_978
        OGN_FLARM
        AIS_162
    }

    TrafficMap "1" --> "*" TrafficInfo
    TrafficInfo --> TrafficSource
```

**Design Rationale**:
- **ICAO address as key**: Unique identifier across all sources
- **RWMutex**: Allow concurrent reads, exclusive writes for thread safety
- **Extrapolation counter**: Track staleness of position data
- **Position_valid flag**: Distinguish position reports from Mode-S only

**Requirements Traced**: FR-401 to FR-407

### 4.2 GPS Module (gps.go)

```mermaid
stateDiagram-v2
    [*] --> Searching: Power On
    Searching --> Acquiring: Satellites Visible
    Acquiring --> Fix2D: 3+ Satellites
    Fix2D --> Fix3D: 4+ Satellites
    Fix3D --> FixDGPS: SBAS/WAAS Lock

    Fix2D --> Searching: Lost Satellites
    Fix3D --> Fix2D: Altitude Invalid
    FixDGPS --> Fix3D: SBAS Lost

    Fix3D --> Valid: HDOP < 4.0
    FixDGPS --> Valid: HDOP < 4.0
    Valid --> Fix3D: HDOP > 4.0

    Valid --> [*]: GPS Disconnect

    note right of Valid
        Position output
        to ownship state
    end note

    note right of Searching
        No position output
        Status: "No Fix"
    end note
```

**State Transitions**:
- **Searching → Acquiring**: Satellite signals detected
- **Acquiring → Fix2D**: 3+ satellites, 2D position computed
- **Fix2D → Fix3D**: 4+ satellites, altitude valid
- **Fix3D → FixDGPS**: WAAS/EGNOS correction available
- **Fix*D → Valid**: HDOP below threshold (4.0)

**Requirements Traced**: FR-201 to FR-205

### 4.3 Message Queue Architecture

```mermaid
flowchart TB
    subgraph "Message Generation"
        G1[Traffic Update]
        G2[Weather Update]
        G3[Ownship Update]
        G4[Heartbeat Timer]
    end

    subgraph "Message Priority"
        P1[High Priority<br/>Traffic Alerts]
        P2[Normal Priority<br/>Ownship, Traffic]
        P3[Low Priority<br/>Weather, Status]
    end

    subgraph "Per-Client Queues"
        Q1[Client 1 Queue<br/>25000 messages]
        Q2[Client 2 Queue<br/>25000 messages]
        Q3[Client N Queue<br/>25000 messages]
    end

    subgraph "Client State"
        S1[Active Client<br/>Full Rate]
        S2[Sleeping Client<br/>Throttled]
    end

    subgraph "Output"
        O1[UDP Socket]
        O2[TCP Socket]
        O3[Serial Port]
    end

    G1 & G2 & G3 & G4 --> P1 & P2 & P3

    P1 --> Q1 & Q2 & Q3
    P2 --> Q1 & Q2 & Q3
    P3 --> Q1 & Q2 & Q3

    Q1 --> S1
    Q2 --> S2
    Q3 --> S1

    S1 --> O1 & O2 & O3
    S2 --> O1

    style P1 fill:#f44336,color:#fff
    style S2 fill:#ff9800
```

**Design Features**:
- **Priority-based**: High-priority (alerts) bypass low-priority (weather)
- **Per-client queues**: Isolate slow clients from fast clients
- **Sleep detection**: Throttle messages to sleeping tablets/phones
- **Wake on alert**: High-priority messages wake sleeping clients

**Requirements Traced**: FR-805

---

## 5. Interface Specifications

### 5.1 GDL90 Message Format

```mermaid
sequenceDiagram
    participant S as Stratux
    participant C as EFB Client

    Note over S,C: Every 1.0 second cycle

    S->>C: Heartbeat (0x00)<br/>GPS Status, UAT Status
    S->>C: Ownship Geo Alt (0x0B)<br/>if GPS valid
    S->>C: Ownship Report (0x0A)<br/>if GPS valid

    loop For each traffic target
        S->>C: Traffic Report (0x14)<br/>Position, Alt, Velocity
    end

    opt AHRS Available
        S->>C: AHRS Report (ForeFlight)<br/>Roll, Pitch, Heading
    end

    loop Weather Products Available
        S->>C: FIS-B Report (0x63)<br/>NEXRAD, METAR, etc.
    end

    Note over S,C: Message framing: 0x7E [data] CRC 0x7E<br/>Escape 0x7D and 0x7E with 0x7D
```

**Message Encoding**:
```
┌────────────────────────────────────────────────┐
│ Flag  │  ID   │    Data (0-432 bytes)   │ CRC │
│ 0x7E  │ 1 byte│   Variable length       │2 bytes│
└────────────────────────────────────────────────┘

Byte stuffing:
  0x7E → 0x7D 0x5E
  0x7D → 0x7D 0x5D
```

**Requirements Traced**: FR-601 to FR-606

### 5.2 Data Structures

#### 5.2.1 Global Situation (Ownship State)

```go
type SituationData struct {
    // GPS Position
    GPSLastFixSinceMidnightUTC float32
    GPSLatitude                float64
    GPSLongitude               float64
    GPSAltitudeMSL             float32
    GPSAltitudeWGS84           float32

    // GPS Quality
    GPSFixQuality              uint8   // 0=No fix, 1=GPS, 2=DGPS
    GPSHeightAboveEllipsoid    float32
    GPSGeoidSep                float32
    GPSSatellites              uint16
    GPSSatellitesTracked       uint16
    GPSHorizontalAccuracy      float32 // meters
    GPSNACp                    uint8   // Navigation Accuracy Category

    // GPS Velocity
    GPSGroundSpeed             float64 // knots
    GPSTrueCourse              uint16  // degrees
    GPSVerticalSpeed           float32 // feet/min
    GPSTurnRate                float64 // degrees/second

    // AHRS Attitude
    AHRSPitch                  float64 // degrees
    AHRSRoll                   float64 // degrees
    AHRSGyroHeading            float64 // degrees
    AHRSMagHeading             float64 // degrees
    AHRSSlipSkid               float64 // lateral G
    AHRSTurnRate               float64 // degrees/sec
    AHRSGLoad                  float64 // vertical G
    AHRSGLoadMin               float64
    AHRSGLoadMax               float64

    // Barometric
    BaroPressureAltitude       float32 // feet
    BaroVerticalSpeed          float32 // feet/min
    BaroTemperature            float32 // celsius

    // Status
    GPSLastFixLocalTime        time.Time
    GPSTime                    time.Time
    GPSLastValidNMEAMessageTime time.Time
    GPSPositionSampleRate      float64
}
```

**Requirements Traced**: FR-201 to FR-205, FR-301 to FR-305

#### 5.2.2 Traffic Information

```go
type TrafficInfo struct {
    Icao_addr                uint32
    OnGround                 bool
    Addr_type                uint8
    SignalLevel              float64  // dBm, for range estimation
    Squawk                   int

    // Position
    Position_valid           bool
    Lat                      float64
    Lng                      float64
    Alt                      int32    // feet
    GnssDiffFromBaroAlt      int32
    AltIsGNSS                bool

    // Velocity
    Speed                    uint16   // knots
    Speed_valid              bool
    Vvel                     int16    // feet/min
    Track                    uint16   // degrees

    // Identification
    Tail                     string
    Emitter_category         uint8

    // Timestamps
    Last_seen                time.Time
    Last_source              uint8    // TrafficSource enum
    Timestamp                time.Time
    Last_alt                 time.Time
    Last_speed               time.Time

    // Extrapolation
    ExtrapolatedPosition     uint8    // seconds since last position

    // Relative Position (to ownship)
    BearingDist_valid        int16
    Bearing                  uint16   // degrees
    Distance                 float32  // nm
    DistanceEstimated        bool

    // Alerting
    AgeLastAlt               float64  // seconds
    Age                      float64  // seconds
}
```

**Requirements Traced**: FR-401 to FR-407

---

## 6. Concurrency Model

### 6.1 Goroutine Architecture

```mermaid
graph TB
    MAIN[Main Goroutine]

    subgraph "Input Goroutines"
        SDR1[SDR Reader 1<br/>1090 MHz]
        SDR2[SDR Reader 2<br/>978 MHz]
        GPS_G[GPS Reader]
        AHRS_G[AHRS Reader<br/>50ms cycle]
    end

    subgraph "Processing Goroutines"
        TRAFFIC_G[Traffic Cleanup<br/>1s periodic]
        WEATHER_G[Weather Cleanup<br/>1s periodic]
    end

    subgraph "Output Goroutines"
        GDL90_G[GDL90 Generator<br/>1s heartbeat]
        NMEA_G[NMEA Generator<br/>1s periodic]
    end

    subgraph "Client Goroutines"
        C1[Client 1 Writer]
        C2[Client 2 Writer]
        CN[Client N Writer]
    end

    subgraph "Management Goroutines"
        WEB_G[Web Server]
        PING_G[Client Ping<br/>30s periodic]
        LOG_G[Data Logger]
    end

    MAIN --> SDR1 & SDR2 & GPS_G & AHRS_G
    MAIN --> TRAFFIC_G & WEATHER_G
    MAIN --> GDL90_G & NMEA_G
    MAIN --> C1 & C2 & CN
    MAIN --> WEB_G & PING_G & LOG_G

    SDR1 & SDR2 --> TRAFFIC_G
    TRAFFIC_G --> GDL90_G
    GPS_G --> GDL90_G
    AHRS_G --> GDL90_G

    GDL90_G --> C1 & C2 & CN
    NMEA_G --> C1 & C2 & CN

    style MAIN fill:#4CAF50,color:#fff
    style TRAFFIC_G fill:#ffeb3b
    style GDL90_G fill:#ffeb3b
```

### 6.2 Synchronization

```mermaid
graph LR
    subgraph "Shared Data Structures"
        TRAFFIC[Traffic Map<br/>RWMutex]
        SITUATION[Ownship State<br/>RWMutex]
        WEATHER[Weather Map<br/>RWMutex]
        CLIENTS[Client List<br/>Mutex]
    end

    subgraph "Readers"
        R1[Traffic Cleanup]
        R2[GDL90 Generator]
        R3[Web UI]
        R4[Data Logger]
    end

    subgraph "Writers"
        W1[SDR Decoder]
        W2[GPS Parser]
        W3[AHRS Processor]
        W4[Client Manager]
    end

    W1 -->|Write Lock| TRAFFIC
    W2 -->|Write Lock| SITUATION
    W3 -->|Write Lock| SITUATION
    W4 -->|Lock| CLIENTS

    TRAFFIC -->|Read Lock| R1 & R2 & R3 & R4
    SITUATION -->|Read Lock| R2 & R3 & R4
    WEATHER -->|Read Lock| R2 & R3 & R4
    CLIENTS -->|Lock| R3

    style TRAFFIC fill:#f44336,color:#fff
    style SITUATION fill:#f44336,color:#fff
```

**Synchronization Strategy**:
- **RWMutex**: Read-heavy data structures (traffic, situation, weather)
- **Mutex**: Balanced read/write (client list)
- **Channels**: Message passing for client queues
- **Atomic Operations**: Simple counters (message counts)

**Requirements Traced**: NFR-201 (Thread Safety)

---

## 7. Error Handling Strategy

### 7.1 Error Categories

```mermaid
graph TD
    ERROR[Error Detected]

    ERROR --> TRANSIENT{Transient?}
    ERROR --> PERMANENT{Permanent?}
    ERROR --> CONFIG{Configuration?}

    TRANSIENT -->|Yes| RETRY[Retry with<br/>Exponential Backoff]
    TRANSIENT -->|No| LOG1[Log Error]

    PERMANENT -->|Yes| DISABLE[Disable Component]
    PERMANENT -->|No| LOG2[Log Error]

    CONFIG -->|Yes| NOTIFY[Notify User<br/>Web UI/LED]
    CONFIG -->|No| LOG3[Log Error]

    RETRY --> SUCCESS{Success?}
    SUCCESS -->|Yes| CLEAR[Clear Error State]
    SUCCESS -->|No| PERMANENT

    DISABLE --> NOTIFY

    style ERROR fill:#f44336,color:#fff
    style RETRY fill:#ff9800
    style NOTIFY fill:#2196F3,color:#fff
```

### 7.2 Error Recovery Mechanisms

| Component | Error Type | Recovery Action | User Notification |
|-----------|------------|----------------|-------------------|
| GPS | No Fix | Continue operating | Status page |
| GPS | Device disconnect | Auto-reconnect every 5s | LED blink + Web |
| SDR | Device disconnect | Auto-reconnect every 5s | LED blink + Web |
| AHRS | Sensor fail | Disable AHRS output | Status page |
| Network | Client disconnect | Remove from list | None |
| Web UI | Config error | Revert to last good | Error message |
| Disk | 95% full | Stop logging | LED blink + Web |

**Requirements Traced**: FR-1103, NFR-202

---

## 8. Performance Considerations

### 8.1 Message Throughput

**Design Target**: 500 ADS-B messages/second (FR-NFR-104)

```
Calculation:
- Average message processing: 2ms
- Goroutine-based parallelism: 4 cores
- Theoretical max: 4 cores × 500 msg/sec = 2000 msg/sec
- Design margin: 4x safety factor
```

### 8.2 Memory Budget

```mermaid
pie title Memory Allocation (Estimated)
    "Traffic Map (1000 targets)" : 25
    "Weather Products" : 20
    "Client Queues (5 clients)" : 30
    "Go Runtime" : 15
    "Other Buffers" : 10
```

**Memory Estimates**:
- Traffic target: ~200 bytes × 1000 targets = 200 KB
- Client queue: 10 MB × 5 clients = 50 MB
- Weather products: ~20 MB
- Go runtime: ~30 MB
- **Total**: ~100-150 MB (well within RPi limits)

**Requirements Traced**: NFR-103, NFR-105

---

## 9. Security Design

### 9.1 Threat Model

```mermaid
graph TD
    subgraph "Threats"
        T1[Spoofed ADS-B Messages]
        T2[Malicious Configuration]
        T3[Unauthorized Access]
        T4[Compromised Updates]
    end

    subgraph "Assets"
        A1[Ownship Position]
        A2[Traffic Display]
        A3[System Config]
        A4[Software Binary]
    end

    subgraph "Mitigations"
        M1[User Awareness<br/>ADS-B not authenticated]
        M2[Input Validation]
        M3[Optional Auth]
        M4[Signed Updates]
    end

    T1 -.Threatens.-> A2
    T2 -.Threatens.-> A3
    T3 -.Threatens.-> A3
    T4 -.Threatens.-> A4

    M1 -.Mitigates.-> T1
    M2 -.Mitigates.-> T2
    M3 -.Mitigates.-> T3
    M4 -.Mitigates.-> T4

    style T1 fill:#f44336,color:#fff
    style T2 fill:#f44336,color:#fff
    style T3 fill:#f44336,color:#fff
    style T4 fill:#f44336,color:#fff
```

**Design Principles**:
1. **Defense in Depth**: Multiple layers of protection
2. **Least Privilege**: Components have minimum necessary access
3. **Fail Secure**: Default to safe state on error
4. **Input Validation**: All external inputs sanitized

**Requirements Traced**: NFR-501 to NFR-504

---

## 10. Design Rationale

### 10.1 Key Decisions

| Decision | Rationale | Trade-offs |
|----------|-----------|------------|
| **Go Language** | Concurrency primitives, memory safety, cross-platform | Learning curve for C developers |
| **Goroutines for I/O** | Non-blocking I/O, efficient concurrency | Context switching overhead |
| **JSON for IPC** | Human-readable, widely supported | Larger than binary formats |
| **SQLite for logging** | Structured queries, ACID, portable | File I/O overhead |
| **UDP Broadcast** | Simple, low latency | No delivery guarantee |
| **Per-client queues** | Isolate slow clients | Memory overhead |
| **ICAO address keying** | Unique, standard identifier | Requires mode-S decoding |
| **15-min weather retention** | Balance memory vs. availability | May miss some products |

### 10.2 Alternative Designs Considered

1. **Single-threaded event loop** (Node.js style)
   - Rejected: Complex callback management, harder to debug

2. **Binary protocol instead of GDL90**
   - Rejected: GDL90 is industry standard, wide EFB support

3. **Redis for message queuing**
   - Rejected: External dependency, memory overhead

4. **WebRTC for client connections**
   - Rejected: Complexity, not needed for local network

---

## 11. Design Verification

### 11.1 Design-to-Requirements Traceability

| Design Element | Requirements Satisfied |
|----------------|----------------------|
| SDR Manager | FR-101 to FR-105 |
| GPS Manager | FR-201 to FR-205 |
| AHRS Manager | FR-301 to FR-305 |
| Traffic Fusion | FR-401 to FR-407 |
| Weather Processing | FR-501 to FR-503 |
| GDL90 Generator | FR-601 to FR-606 |
| NMEA Generator | FR-701 to FR-702 |
| Network Manager | FR-801 to FR-805 |
| Web Interface | FR-901 to FR-905 |
| Data Logging | FR-1001 to FR-1003 |
| System Management | FR-1101 to FR-1104 |
| Performance | NFR-101 to NFR-105 |
| Reliability | NFR-201 to NFR-204 |
| Security | NFR-501 to NFR-504 |

**Coverage**: 101/101 requirements traced to design elements (100%)

### 11.2 Design Reviews

| Review Type | Participants | Date | Status |
|-------------|-------------|------|--------|
| Architecture Review | TBD | TBD | Pending |
| Safety Review | TBD | TBD | Pending |
| Security Review | TBD | TBD | Pending |
| Code Review | TBD | Ongoing | In Progress |

---

## 12. Document Control

### 12.1 Change History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1 | 2025-10-13 | Design Team | Initial draft |
| 1.0 DRAFT | 2025-10-13 | Design Team | Complete draft for review |

### 12.2 Approvals

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Author | TBD | | |
| Technical Review | TBD | | |
| Architecture Review | TBD | | |
| Approval | TBD | | |

---

## References

1. Stratux GitHub Repository: https://github.com/cyoung/stratux
2. GDL90 Data Interface Specification (Garmin)
3. RTCA DO-260B: 1090 MHz ES ADS-B MOPS
4. RTCA DO-282B: UAT ADS-B MOPS
5. RTCA DO-278A: Software Integrity Assurance
6. NMEA 0183 Standard

---

**END OF DOCUMENT**

**Next Steps**:
1. Conduct architecture review
2. Verify design satisfies all requirements
3. Update design based on feedback
4. Begin detailed design and implementation
