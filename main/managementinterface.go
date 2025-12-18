/*
	Copyright (c) 2015-2016 Christopher Young
	Distributable under the terms of The "BSD New" License
	that can be found in the LICENSE file, herein included
	as part of this header.

	managementinterface.go: Web interfaces (JSON and websocket), web server for web interface HTML.
*/

package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	humanize "github.com/dustin/go-humanize"
	"golang.org/x/net/websocket"

	"github.com/stratux/stratux/common"
)

// varLogDir is kept for backward compatibility, points to varLogDirPath
// The actual path is now configurable via varLogDirPath in config_paths.go

type SettingMessage struct {
	Setting string `json:"setting"`
	Value   bool   `json:"state"`
}

type MbTileConnectionCacheEntry struct {
	Path     string
	Conn     *sql.DB
	Metadata map[string]string
	fileTime time.Time
}

func (this *MbTileConnectionCacheEntry) IsOutdated() bool {
	file, err := os.Stat(this.Path)
	if err != nil {
		return true
	}
	modTime := file.ModTime()
	return modTime != this.fileTime
}

func NewMbTileConnectionCacheEntry(path string, conn *sql.DB) *MbTileConnectionCacheEntry {
	file, err := os.Stat(path)
	if err != nil {
		return nil
	}
	return &MbTileConnectionCacheEntry{path, conn, nil, file.ModTime()}
}

var mbtileCacheLock = sync.Mutex{}
var mbtileConnectionCache = make(map[string]MbTileConnectionCacheEntry)

// Weather updates channel.
var weatherUpdate *uibroadcaster
var trafficUpdate *uibroadcaster
var radarUpdate *uibroadcaster
var gdl90Update *uibroadcaster

func handleGDL90WS(conn *websocket.Conn) {
	// Subscribe the socket to receive updates.
	gdl90Update.AddSocket(conn)

	// Connection closes when function returns. Since uibroadcast is writing and we don't need to read anything (for now), just keep it busy.
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			break
		}
		if buf[0] != 0 { // Dummy.
			continue
		}
		time.Sleep(1 * time.Second)
	}
}

// Situation updates channel.
var situationUpdate *uibroadcaster

// Raw weather (UATFrame packet stream) update channel.
var weatherRawUpdate *uibroadcaster

/*
The /weather websocket starts off by sending the current buffer of weather messages, then sends updates as they are received.
*/
func handleWeatherWS(conn *websocket.Conn) {
	// Subscribe the socket to receive updates.
	weatherUpdate.AddSocket(conn)

	// Connection closes when function returns. Since uibroadcast is writing and we don't need to read anything (for now), just keep it busy.
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			break
		}
		if buf[0] != 0 { // Dummy.
			continue
		}
		time.Sleep(1 * time.Second)
	}
}

func handleJsonIo(conn *websocket.Conn) {
	trafficMutex.Lock()
	for _, traf := range traffic {
		if !traf.Position_valid { // Don't send unless a valid position exists.
			continue
		}
		trafficJSON, err := json.Marshal(&traf)
		if err != nil {
			log.Printf("Error marshaling traffic JSON: %s", err.Error())
			continue
		}
		conn.Write(trafficJSON)
	}
	// Subscribe the socket to receive updates.
	trafficUpdate.AddSocket(conn)
	radarUpdate.AddSocket(conn)
	weatherRawUpdate.AddSocket(conn)
	situationUpdate.AddSocket(conn)

	trafficMutex.Unlock()

	// Connection closes when function returns. Since uibroadcast is writing and we don't need to read anything (for now), just keep it busy.
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			break
		}
		if buf[0] != 0 { // Dummy.
			continue
		}
		time.Sleep(1 * time.Second)
	}
}

// Works just as weather updates do.

func handleTrafficWS(conn *websocket.Conn) {
	trafficMutex.Lock()
	for _, traf := range traffic {
		if !traf.Position_valid { // Don't send unless a valid position exists.
			continue
		}
		trafficJSON, err := json.Marshal(&traf)
		if err != nil {
			log.Printf("Error marshaling traffic JSON: %s", err.Error())
			continue
		}
		conn.Write(trafficJSON)
	}
	// Subscribe the socket to receive updates.
	trafficUpdate.AddSocket(conn)
	trafficMutex.Unlock()

	// Connection closes when function returns. Since uibroadcast is writing and we don't need to read anything (for now), just keep it busy.
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			break
		}
		if buf[0] != 0 { // Dummy.
			continue
		}
		time.Sleep(1 * time.Second)
	}
}

func handleRadarWS(conn *websocket.Conn) {
	trafficMutex.Lock()
	// Subscribe the socket to receive updates. Not necessary to send old traffic
	radarUpdate.AddSocket(conn)
	trafficMutex.Unlock()

	radarUpdate.SendJSON(globalSettings)

	// Connection closes when function returns. Since uibroadcast is writing and we don't need to read anything (for now), just keep it busy.
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			break
		}
		if buf[0] != 0 { // Dummy.
			continue
		}
		time.Sleep(1 * time.Second)
	}
}

func handleStatusWS(conn *websocket.Conn) {
	//	log.Printf("Web client connected.\n")

	timer := time.NewTicker(1 * time.Second)
	for {
		// The below is not used, but should be if something needs to be streamed from the web client ever in the future.
		/*		var msg SettingMessage
				err := websocket.JSON.Receive(conn, &msg)
				if err == io.EOF {
					break
				} else if err != nil {
					log.Printf("handleStatusWS: %s\n", err.Error())
				} else {
					// Use 'msg'.
				}
		*/

		// Send status.
		update, err := json.Marshal(&globalStatus)
		if err != nil {
			log.Printf("Error marshaling status JSON: %s", err.Error())
			continue
		}
		_, err = conn.Write(update)

		if err != nil {
			//			log.Printf("Web client disconnected.\n")
			break
		}
		<-timer.C
	}
}

func handleSituationWS(conn *websocket.Conn) {
	timer := time.NewTicker(100 * time.Millisecond)
	for {
		situationJSON, err := json.Marshal(&mySituation)
		if err != nil {
			log.Printf("Error marshaling situation JSON: %s", err.Error())
			continue
		}
		_, err = conn.Write(situationJSON)

		if err != nil {
			break
		}
		<-timer.C

	}

}

// AJAX call - /getStatus. Responds with current global status
// a webservice call for the same data available on the websocket but when only a single update is needed
func handleStatusRequest(w http.ResponseWriter, _ *http.Request) {
	setJSONHeadersWithNoCache(w)
	statusJSON, err := json.Marshal(&globalStatus)
	if err != nil {
		log.Printf("Error marshaling status JSON: %s\n", err.Error())
	}
	fmt.Fprintf(w, "%s\n", statusJSON)
}

// AJAX call - /getSituation. Responds with current situation (lat/lon/gdspeed/track/pitch/roll/heading/etc.)
func handleSituationRequest(w http.ResponseWriter, _ *http.Request) {
	setJSONHeadersWithNoCache(w)
	situationJSON, err := json.Marshal(&mySituation)
	if err != nil {
		log.Printf("Error marshaling situation JSON: %s\n", err.Error())
	}
	fmt.Fprintf(w, "%s\n", situationJSON)
}

// AJAX call - /getTowers. Responds with all ADS-B ground towers that have sent messages that we were able to parse, along with its stats.
func handleTowersRequest(w http.ResponseWriter, _ *http.Request) {
	setJSONHeadersWithNoCache(w)

	ADSBTowerMutex.Lock()
	towersJSON, err := json.Marshal(&ADSBTowers)
	if err != nil {
		log.Printf("Error sending tower JSON data: %s\n", err.Error())
	}
	// for testing purposes, we can return a fixed reply
	// towersJSON = []byte(`{"(38.490880,-76.135554)":{"Lat":38.49087953567505,"Lng":-76.13555431365967,"Signal_strength_last_minute":100,"Signal_strength_max":67,"Messages_last_minute":1,"Messages_total":1059},"(38.978698,-76.309276)":{"Lat":38.97869825363159,"Lng":-76.30927562713623,"Signal_strength_last_minute":495,"Signal_strength_max":32,"Messages_last_minute":45,"Messages_total":83},"(39.179285,-76.668413)":{"Lat":39.17928457260132,"Lng":-76.66841268539429,"Signal_strength_last_minute":50,"Signal_strength_max":24,"Messages_last_minute":1,"Messages_total":16},"(39.666309,-74.315300)":{"Lat":39.66630935668945,"Lng":-74.31529998779297,"Signal_strength_last_minute":9884,"Signal_strength_max":35,"Messages_last_minute":4,"Messages_total":134}}`)
	fmt.Fprintf(w, "%s\n", towersJSON)
	ADSBTowerMutex.Unlock()
}

// AJAX call - /getSatellites. Responds with all GNSS satellites that are being tracked, along with status information.
func handleSatellitesRequest(w http.ResponseWriter, _ *http.Request) {
	setJSONHeadersWithNoCache(w)
	mySituation.muSatellite.Lock()
	satellitesJSON, err := json.Marshal(&Satellites)
	if err != nil {
		log.Printf("Error sending GNSS satellite JSON data: %s\n", err.Error())
	}
	fmt.Fprintf(w, "%s\n", satellitesJSON)
	mySituation.muSatellite.Unlock()
}

// AJAX call - /getSettings. Responds with all stratux.conf data.
func handleSettingsGetRequest(w http.ResponseWriter, _ *http.Request) {
	setJSONHeadersWithNoCache(w)
	settingsJSON, err := json.Marshal(&globalSettings)
	if err != nil {
		log.Printf("%s", err)
	}
	fmt.Fprintf(w, "%s\n", settingsJSON)
}

func handleRegionGet(w http.ResponseWriter, _ *http.Request) {
	setJSONHeadersWithNoCache(w)
	switch globalSettings.RegionSelected {
	case 1:
		RegionSettings.IsSet = true
		RegionSettings.Region = "US"
	case 2:
		RegionSettings.IsSet = true
		RegionSettings.Region = "EU"
	default:
		RegionSettings.IsSet = false
	}

	regionJSON, err := json.Marshal(&RegionSettings)
	if err != nil {
		log.Printf("%s", err)
	}
	fmt.Fprintf(w, "%s\n", regionJSON)
}

// AJAX call - /setRegion. receives via POST command, any/all stratux.conf data.
func handleRegionSet(w http.ResponseWriter, r *http.Request) {
	// define header in support of cross-domain AJAX
	setJSONHeadersWithNoCache(w)
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	if r.Method == "POST" {
		// raw, _ := httputil.DumpRequest(r, true)
		// log.Printf("handleRegionSet:raw: %s\n", raw)

		decoder := json.NewDecoder(r.Body)
		for {
			var msg map[string]interface{} // support arbitrary JSON

			err := decoder.Decode(&msg)
			if err == io.EOF {
				break
			} else if err != nil {
				log.Printf("handleRegionSet:error: %s\n", err.Error())
			} else {
				for key, val := range msg {
					// log.Printf("handleRegionSet:json: testing for key:%s of type %s\n", key, reflect.TypeOf(val))
					switch key {
					case "Region":
						regionStr := val.(string)
						log.Printf("String is %s\n", regionStr)
						if regionStr == "US" {
							globalSettings.RegionSelected = 1
						} else if regionStr == "EU" {
							globalSettings.RegionSelected = 2
						} else {
							globalSettings.RegionSelected = 0
						}
						changeRegionSettings()
					default:
						log.Printf("handleRegionSet:json: unrecognized key:%s\n", key)
					}
				}
				//				saveSettings()
			}
		}
	}

}

// settingResult holds the side effects from applying a setting.
type settingResult struct {
	reconfigureTracker    bool
	reconfigureFancontrol bool
}

// settingHandler is a function that applies a setting value and returns any side effects.
type settingHandler func(val interface{}) settingResult

// noSideEffects is a convenience for handlers with no side effects.
var noSideEffects = settingResult{}

// applyOwnshipModeS parses and validates Mode S codes.
func applyOwnshipModeS(val interface{}) settingResult {
	codes := strings.Split(val.(string), ",")
	codesFinal := make([]string, 0)
	for _, code := range codes {
		code = strings.Trim(code, " ")
		// Expecting a hex string less than 6 characters (24 bits) long.
		if len(code) > 6 { // Too long.
			continue
		}
		// Pad string, must be 6 characters long.
		vals := strings.ToUpper(code)
		for len(vals) < 6 {
			vals = "0" + vals
		}
		hexn, err := hex.DecodeString(vals)
		if err != nil { // Number not valid.
			log.Printf("handleSettingsSetRequest:OwnshipModeS: %s\n", err.Error())
			continue
		}
		codesFinal = append(codesFinal, fmt.Sprintf("%02X%02X%02X", hexn[0], hexn[1], hexn[2]))
	}
	globalSettings.OwnshipModeS = strings.Join(codesFinal, ",")
	return noSideEffects
}

// ipValidationRegex is compiled once for efficiency.
var ipValidationRegex = regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)

// applyStaticIps parses and validates static IP addresses.
func applyStaticIps(val interface{}) settingResult {
	ipsStr := val.(string)
	ips := strings.Split(ipsStr, " ")
	if ipsStr == "" {
		ips = make([]string, 0)
	}

	errMsg := ""
	for _, ip := range ips {
		// Verify IP format
		if !ipValidationRegex.MatchString(ip) {
			errMsg = errMsg + "Invalid IP: " + ip + ". "
		}
	}
	if errMsg != "" {
		log.Printf("handleSettingsSetRequest:StaticIps: %s\n", errMsg)
		return noSideEffects // Don't update on validation error
	}
	globalSettings.StaticIps = ips
	return noSideEffects
}

// applyWiFiClientNetworks parses WiFi client network configurations.
func applyWiFiClientNetworks(val interface{}) settingResult {
	var networks = make([]wifiClientNetwork, 0)
	for _, rawNetwork := range val.([]interface{}) {
		network := rawNetwork.(map[string]interface{})
		networks = append(networks, wifiClientNetwork{network["SSID"].(string), network["Password"].(string)})
	}
	setWifiClientNetworks(networks)
	return noSideEffects
}

// applyBaudRate updates baud rate for all serial outputs.
func applyBaudRate(val interface{}) settingResult {
	if globalSettings.SerialOutputs != nil {
		for dev, serialOut := range globalSettings.SerialOutputs {
			newBaud := int(val.(float64))
			if newBaud == serialOut.Baud { // Same baud rate. No change.
				continue
			}
			log.Printf("changing %s baud rate from %d to %d.\n", dev, serialOut.Baud, newBaud)
			serialOut.Baud = newBaud
			globalSettings.SerialOutputs[dev] = serialOut
			closeSerial(dev)
		}
	}
	return noSideEffects
}

// settingHandlers maps setting keys to their handler functions.
// This eliminates the need for switch statements and makes adding new settings trivial.
var settingHandlers = map[string]settingHandler{
	// Simple boolean settings (no side effects)
	"DarkMode":              func(v interface{}) settingResult { globalSettings.DarkMode = v.(bool); return noSideEffects },
	"UAT_Enabled":           func(v interface{}) settingResult { globalSettings.UAT_Enabled = v.(bool); return noSideEffects },
	"ES_Enabled":            func(v interface{}) settingResult { globalSettings.ES_Enabled = v.(bool); return noSideEffects },
	"OGN_Enabled":           func(v interface{}) settingResult { globalSettings.OGN_Enabled = v.(bool); return noSideEffects },
	"AIS_Enabled":           func(v interface{}) settingResult { globalSettings.AIS_Enabled = v.(bool); return noSideEffects },
	"APRS_Enabled":          func(v interface{}) settingResult { globalSettings.APRS_Enabled = v.(bool); return noSideEffects },
	"Ping_Enabled":          func(v interface{}) settingResult { globalSettings.Ping_Enabled = v.(bool); return noSideEffects },
	"Pong_Enabled":          func(v interface{}) settingResult { globalSettings.Pong_Enabled = v.(bool); return noSideEffects },
	"OGNI2CTXEnabled":       func(v interface{}) settingResult { globalSettings.OGNI2CTXEnabled = v.(bool); return noSideEffects },
	"GPS_Enabled":           func(v interface{}) settingResult { globalSettings.GPS_Enabled = v.(bool); return noSideEffects },
	"DEBUG":                 func(v interface{}) settingResult { globalSettings.DEBUG = v.(bool); return noSideEffects },
	"DisplayTrafficSource":  func(v interface{}) settingResult { globalSettings.DisplayTrafficSource = v.(bool); return noSideEffects },
	"TraceLog":              func(v interface{}) settingResult { globalSettings.TraceLog = v.(bool); return noSideEffects },
	"AHRSLog":               func(v interface{}) settingResult { globalSettings.AHRSLog = v.(bool); return noSideEffects },
	"EstimateBearinglessDist": func(v interface{}) settingResult { globalSettings.EstimateBearinglessDist = v.(bool); return noSideEffects },

	// Boolean settings with side effects
	"IMU_Sensor_Enabled": func(v interface{}) settingResult {
		globalSettings.IMU_Sensor_Enabled = v.(bool)
		if !globalSettings.IMU_Sensor_Enabled && globalStatus.IMUConnected {
			myIMUReader.Close()
			globalStatus.IMUConnected = false
		}
		return noSideEffects
	},
	"BMP_Sensor_Enabled": func(v interface{}) settingResult {
		globalSettings.BMP_Sensor_Enabled = v.(bool)
		if !globalSettings.BMP_Sensor_Enabled && globalStatus.BMPConnected {
			myPressureReader.Close()
			globalStatus.BMPConnected = false
		}
		return noSideEffects
	},
	"ReplayLog": func(v interface{}) settingResult {
		newVal := v.(bool)
		if newVal != globalSettings.ReplayLog { // Don't mark the files unless there is a change.
			globalSettings.ReplayLog = newVal
		}
		return noSideEffects
	},
	"PersistentLogging": func(v interface{}) settingResult {
		globalSettings.PersistentLogging = v.(bool)
		setPersistentLogging(globalSettings.PersistentLogging)
		return noSideEffects
	},
	"IMUMapping": func(v interface{}) settingResult {
		if globalSettings.IMUMapping != v.([2]int) {
			globalSettings.IMUMapping = v.([2]int)
			myIMUReader.Close()
			globalStatus.IMUConnected = false // Force a restart of the IMU reader
		}
		return noSideEffects
	},

	// Numeric settings
	"Dump1090Gain":   func(v interface{}) settingResult { globalSettings.Dump1090Gain = v.(float64); return noSideEffects },
	"PPM":            func(v interface{}) settingResult { globalSettings.PPM = int(v.(float64)); return noSideEffects },
	"AltitudeOffset": func(v interface{}) settingResult { globalSettings.AltitudeOffset = int(v.(float64)); return noSideEffects },
	"RadarLimits": func(v interface{}) settingResult {
		globalSettings.RadarLimits = int(v.(float64))
		radarUpdate.SendJSON(globalSettings)
		return noSideEffects
	},
	"RadarRange": func(v interface{}) settingResult {
		globalSettings.RadarRange = int(v.(float64))
		radarUpdate.SendJSON(globalSettings)
		return noSideEffects
	},
	"PWMDutyMin": func(v interface{}) settingResult {
		globalSettings.PWMDutyMin = int(v.(float64))
		return settingResult{reconfigureFancontrol: true}
	},

	// String settings
	"WatchList": func(v interface{}) settingResult { globalSettings.WatchList = v.(string); return noSideEffects },
	"GLimits":   func(v interface{}) settingResult { globalSettings.GLimits = v.(string); return noSideEffects },

	// Complex validation handlers
	"OwnshipModeS": applyOwnshipModeS,
	"StaticIps":    applyStaticIps,
	"Baud":         applyBaudRate,

	// WiFi settings
	"WiFiCountry":                   func(v interface{}) settingResult { setWifiCountry(v.(string)); return noSideEffects },
	"WiFiSSID":                      func(v interface{}) settingResult { setWifiSSID(v.(string)); return noSideEffects },
	"WiFiChannel":                   func(v interface{}) settingResult { setWifiChannel(int(v.(float64))); return noSideEffects },
	"WiFiSecurityEnabled":           func(v interface{}) settingResult { setWifiSecurityEnabled(v.(bool)); return noSideEffects },
	"WiFiPassphrase":                func(v interface{}) settingResult { setWifiPassphrase(v.(string)); return noSideEffects },
	"WiFiIPAddress":                 func(v interface{}) settingResult { setWifiIPAddress(v.(string)); return noSideEffects },
	"WiFiMode":                      func(v interface{}) settingResult { setWiFiMode(int(v.(float64))); return noSideEffects },
	"WiFiDirectPin":                 func(v interface{}) settingResult { setWifiDirectPin(v.(string)); return noSideEffects },
	"WiFiClientNetworks":            applyWiFiClientNetworks,
	"WiFiInternetPassThroughEnabled": func(v interface{}) settingResult { setWifiInternetPassthroughEnabled(v.(bool)); return noSideEffects },

	// OGN settings (all trigger tracker reconfiguration)
	"OGNAddrType": func(v interface{}) settingResult { globalSettings.OGNAddrType = int(v.(float64)); return settingResult{reconfigureTracker: true} },
	"OGNAddr":     func(v interface{}) settingResult { globalSettings.OGNAddr = v.(string); return settingResult{reconfigureTracker: true} },
	"OGNAcftType": func(v interface{}) settingResult { globalSettings.OGNAcftType = int(v.(float64)); return settingResult{reconfigureTracker: true} },
	"OGNPilot":    func(v interface{}) settingResult { globalSettings.OGNPilot = v.(string); return settingResult{reconfigureTracker: true} },
	"OGNReg":      func(v interface{}) settingResult { globalSettings.OGNReg = v.(string); return settingResult{reconfigureTracker: true} },
	"OGNTxPower":  func(v interface{}) settingResult { globalSettings.OGNTxPower = int(v.(float64)); return settingResult{reconfigureTracker: true} },
}

// AJAX call - /setSettings. receives via POST command, any/all stratux.conf data.
func handleSettingsSetRequest(w http.ResponseWriter, r *http.Request) {
	// define header in support of cross-domain AJAX
	setJSONHeadersWithNoCache(w)
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	// for an OPTION method request, we return header without processing.
	// this insures we are recognized as supporting cross-domain AJAX REST calls
	if r.Method == "POST" {
		// raw, _ := httputil.DumpRequest(r, true)
		// log.Printf("handleSettingsSetRequest:raw: %s\n", raw)

		decoder := json.NewDecoder(r.Body)
		for {
			var msg map[string]interface{} // support arbitrary JSON

			err := decoder.Decode(&msg)
			if err == io.EOF {
				break
			} else if err != nil {
				log.Printf("handleSettingsSetRequest:error: %s\n", err.Error())
			} else {
				reconfigureTracker := false
				reconfigureFancontrol := false
				for key, val := range msg {
					// log.Printf("handleSettingsSetRequest:json: testing for key:%s of type %s\n", key, reflect.TypeOf(val))
					if handler, ok := settingHandlers[key]; ok {
						result := handler(val)
						reconfigureTracker = reconfigureTracker || result.reconfigureTracker
						reconfigureFancontrol = reconfigureFancontrol || result.reconfigureFancontrol
					} else {
						log.Printf("handleSettingsSetRequest:json: unrecognized key:%s\n", key)
					}
				}
				saveSettings()
				applyNetworkSettings(false, false)
				if reconfigureTracker && detectedTracker != nil {
					writeTrackerConfigFromSettings()
				}
				if reconfigureFancontrol {
					exec.Command("killall", "-SIGUSR1", "fancontrol").Run()
				}
			}
		}

		// while it may be redundant, we return the latest settings
		settingsJSON, err := json.Marshal(&globalSettings)
		if err != nil {
			log.Printf("Error marshaling settings JSON: %s", err.Error())
		}
		fmt.Fprintf(w, "%s\n", settingsJSON)
	}
}

func setPersistentLogging(persistent bool) {
	if persistent {
		overlayctl("disable")
	} else {
		overlayctl("enable")
	}
}

func handleShutdownRequest(w http.ResponseWriter, r *http.Request) {
	syscall.Sync()
	exec.Command("systemctl", "poweroff").Run()
}

// doReboot is a function variable for rebooting the system.
// It can be replaced in tests to mock system calls.
var doReboot = func() {
	syscall.Sync()
	exec.Command("systemctl", "reboot").Run()
}

func handleDeleteLogFile(_ http.ResponseWriter, _ *http.Request) {
	log.Println("Clearing debug log file")
	clearDebugLogFile()
}

func handleDeleteAHRSLogFiles(w http.ResponseWriter, _ *http.Request) {
	files, err := os.ReadDir(varLogDirPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting AHRS logs: %s", err), http.StatusNotFound)
		return
	}

	var fn string
	for _, f := range files {
		fn = f.Name()
		if v, _ := filepath.Match("sensors_*.csv", fn); v {
			filePath := filepath.Join(varLogDirPath, fn)
			if err := os.Remove(filePath); err != nil {
				log.Printf("Error deleting AHRS log file %s: %s\n", fn, err.Error())
			} else {
				log.Printf("Deleted AHRS log file %s\n", fn)
			}
		}
		analysisLogger = nil
	}
}

func handleDevelModeToggle(_ http.ResponseWriter, _ *http.Request) {
	log.Println("Enabling developer mode")
	globalSettings.DeveloperMode = true
	saveSettings()
}

func handleRestartRequest(_ http.ResponseWriter, _ *http.Request) {
	log.Printf("handleRestartRequest called\n")
	go doRestartApp()
}

func handleRebootRequest(w http.ResponseWriter, r *http.Request) {
	setJSONHeadersWithNoCache(w)
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	go delayReboot()
}

func handleOrientAHRS(w http.ResponseWriter, r *http.Request) {
	// define header in support of cross-domain AJAX
	setNoCache(w)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	// For an OPTION method request, we return header without processing.
	// This ensures we are recognized as supporting cross-domain AJAX REST calls.
	if r.Method == "POST" {
		var (
			action []byte = make([]byte, 1)
			err    error
		)

		if _, err = r.Body.Read(action); err != nil {
			log.Println("AHRS Error: handleOrientAHRS received invalid request")
			http.Error(w, "orientation received invalid request", http.StatusBadRequest)
		}

		switch action[0] {
		case 'f': // Set sensor "forward" direction (toward nose of airplane).
			f, err := getMinAccelDirection()
			if err != nil {
				log.Printf("AHRS Error: sensor orientation: couldn't read accelerometer: %s\n", err)
				http.Error(w, fmt.Sprintf("couldn't read accelerometer: %s\n", err), http.StatusBadRequest)
				return
			}
			log.Printf("AHRS Info: sensor orientation success! forward axis is %d\n", f)
			globalSettings.IMUMapping = [2]int{f, 0}
		case 'd': // Set sensor "up" direction (toward top of airplane).
			globalSettings.SensorQuaternion = [4]float64{0, 0, 0, 0}
			saveSettings()
			myIMUReader.Close()
			globalStatus.IMUConnected = false // restart the processes depending on the orientation
			ResetAHRSGLoad()
			time.Sleep(2000 * time.Millisecond)
		}
	}
}

func handleCageAHRS(w http.ResponseWriter, r *http.Request) {
	// define header in support of cross-domain AJAX
	setNoCache(w)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	// For an OPTION method request, we return header without processing.
	// This ensures we are recognized as supporting cross-domain AJAX REST calls.
	if r.Method == "POST" {
		CageAHRS()
	}
}

func handleCalibrateAHRS(w http.ResponseWriter, r *http.Request) {
	// define header in support of cross-domain AJAX
	setNoCache(w)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	// For an OPTION method request, we return header without processing.
	// This ensures we are recognized as supporting cross-domain AJAX REST calls.
	if r.Method == "POST" {
		CalibrateAHRS()
	}
}

func handleResetGMeter(w http.ResponseWriter, r *http.Request) {
	// define header in support of cross-domain AJAX
	setNoCache(w)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")

	// For an OPTION method request, we return header without processing.
	// This ensures we are recognized as supporting cross-domain AJAX REST calls.
	if r.Method == "POST" {
		ResetAHRSGLoad()
	}
}

func doRestartApp() {
	time.Sleep(1)
	syscall.Sync()
	out, err := exec.Command("/bin/systemctl", "restart", "stratux").Output()
	if err != nil {
		log.Printf("restart error: %s\n%s", err.Error(), out)
	} else {
		log.Printf("restart: %s\n", out)
	}
}

// AJAX call - /getClients. Responds with all connected clients.
func handleClientsGetRequest(w http.ResponseWriter, r *http.Request) {
	setJSONHeadersWithNoCache(w)
	netMutex.Lock()
	clientsJSON, err := json.Marshal(&clientConnections)
	netMutex.Unlock()
	if err != nil {
		log.Printf("Error marshaling clients JSON: %s", err.Error())
	}
	fmt.Fprintf(w, "%s\n", clientsJSON)
}

// delayReboot is a function variable for delayed system reboot.
// It can be replaced in tests to mock system calls.
var delayReboot = func() {
	time.Sleep(1 * time.Second)
	doReboot()
}

func handleDownloadLogRequest(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=stratux.log")
	http.ServeFile(w, r, filepath.Join(varLogDirPath, "stratux.log"))
}

func handleDownloadAHRSLogsRequest(w http.ResponseWriter, _ *http.Request) {
	// Common error handler
	httpErr := func(w http.ResponseWriter, e error) {
		http.Error(w, fmt.Sprintf("error zipping AHRS logs: %s", e), http.StatusNotFound)
	}

	files, err := os.ReadDir(varLogDirPath)
	if err != nil {
		httpErr(w, err)
		return
	}

	z := zip.NewWriter(w)
	defer z.Close()

	for _, f := range files {
		fn := f.Name()
		v1, _ := filepath.Match("sensors_*.csv", fn)
		v2, _ := filepath.Match("stratux.log", fn)
		if !(v1 || v2) {
			continue
		}

		unzippedFile, err := os.Open(filepath.Join(varLogDirPath, fn))
		if err != nil {
			httpErr(w, err)
			return
		}
		defer unzippedFile.Close()

		finfo, err := f.Info()
		if err != nil {
			httpErr(w, err)
			return
		}
		fh, err := zip.FileInfoHeader(finfo)
		if err != nil {
			httpErr(w, err)
			return
		}
		zippedFile, err := z.CreateHeader(fh)
		if err != nil {
			httpErr(w, err)
			return
		}

		_, err = io.Copy(zippedFile, unzippedFile)
		if err != nil {
			httpErr(w, err)
			return
		}
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=ahrs_logs.zip")
}

func handleDownloadDBRequest(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=stratux.sqlite")
	http.ServeFile(w, r, filepath.Join(varLogDirPath, "stratux.sqlite"))
}

// Upload an update file.
func handleUpdatePostRequest(w http.ResponseWriter, r *http.Request) {
	setJSONHeadersWithNoCache(w)
	overlayctl("unlock")
	reader, err := r.MultipartReader()
	if err != nil {
		log.Printf("Update failed from %s (%s).\n", r.RemoteAddr, sanitizeLogString(err.Error()))
		return
	}

	var temp_filename string
	var upload_filename string

	var base_dir string

	if common.IsRunningAsRoot() {
		base_dir = "/overlay/robase/root"
	} else {
		base_dir = "."
		log.Printf("not running as root, using base_dir of %s", base_dir)
	}

	for {
		part, err := reader.NextPart()
		if err != nil {
			log.Printf("Update failed from %s (%s).\n", r.RemoteAddr, err.Error())
			return
		}
		if part == nil {
			return
		}

		if part.FormName() != "update_file" {
			continue
		}

		temp_filename = fmt.Sprintf("%s/TMP_%s", base_dir, part.FileName())
		upload_filename = fmt.Sprintf("%s/%s", base_dir, part.FileName())

		fi, err := os.OpenFile(temp_filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Printf("Update failed from %s (%s).\n", r.RemoteAddr, err.Error())
			return
		}
		defer fi.Close()
		_, err = io.Copy(fi, part)
		if err != nil {
			log.Printf("Update failed from %s (%s).\n", r.RemoteAddr, err.Error())
			return
		}

		break
	}

	os.Rename(temp_filename, upload_filename)
	log.Printf("%s uploaded %s for update.\n", r.RemoteAddr, upload_filename)
	overlayctl("disable")
	// Successful update upload. Now reboot.
	go delayReboot()
}

// Upload an update file for Pong
func handlePongUpdatePostRequest(w http.ResponseWriter, r *http.Request) {
	setJSONHeadersWithNoCache(w)
	log.Printf("request: %s\n", sanitizeLogString(r.URL.RequestURI()))
	err := r.ParseMultipartForm(8 << 20)
	if err != nil {
		log.Printf("Step 1 Update failed from %s (%s).\n", r.RemoteAddr, sanitizeLogString(err.Error()))
		return
	}
	file, _, err := r.FormFile("pong_update_file")
	if err != nil {
		log.Printf("FormFile returned error %s\n", sanitizeLogString(err.Error()))
		return
	}
	fi, err := os.OpenFile("/tmp/update_pong.zip", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("Cannot open file for saving (%s)\n", err.Error())
		return
	}
	defer fi.Close()
	_, err = io.Copy(fi, file)
	if err != nil {
		log.Printf("Could not copy file (%s)\n", err.Error())
		return
	}
	log.Printf("Set update mode flag to signal Pong to run the update\n")
	pongSetUpdateMode()

	file.Close()
}

func setNoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func setJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
}

// setJSONHeadersWithNoCache combines setNoCache and setJSONHeaders for JSON API endpoints.
// This is the common pattern for most API handlers.
func setJSONHeadersWithNoCache(w http.ResponseWriter) {
	setNoCache(w)
	setJSONHeaders(w)
}

// sanitizeLogString removes or replaces characters that could be used for log injection attacks.
// This prevents attackers from injecting fake log entries via user-controlled input.
func sanitizeLogString(s string) string {
	// Replace newlines and carriage returns to prevent log forging
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	// Replace tabs to maintain log format consistency
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func defaultServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=360") // 5 min, so that if user installs update, he will revalidate soon enough
	//	setNoCache(w)
	http.FileServer(http.Dir(STRATUX_WWW_DIR)).ServeHTTP(w, r)
}

func handleroPartitionRebuild(w http.ResponseWriter, r *http.Request) {
	out, err := exec.Command("/usr/sbin/rebuild_ro_part.sh").Output()

	if err != nil {
		addSingleSystemErrorf("partition-rebuild", "Rebuild RO Partition error: %s", err.Error())
	} else {
		addSingleSystemErrorf("partition-rebuild", "Rebuild RO Partition success: %s", out)
	}

}

// https://gist.github.com/alexisrobert/982674.
// Copyright (c) 2010-2014 Alexis ROBERT <alexis.robert@gmail.com>.
const dirlisting_tpl = `<?xml version="1.0" encoding="iso-8859-1"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
<!-- Modified from lighttpd directory listing -->
<head>
<title>Index of {{.Name}}</title>
<style type="text/css">
a, a:active {text-decoration: none; color: blue;}
a:visited {color: #48468F;}
a:hover, a:focus {text-decoration: underline; color: red;}
body {background-color: #F5F5F5;}
h2 {margin-bottom: 12px;}
table {margin-left: 12px;}
th, td { font: 90% monospace; text-align: left;}
th { font-weight: bold; padding-right: 14px; padding-bottom: 3px;}
td {padding-right: 14px;}
td.s, th.s {text-align: right;}
div.list { background-color: white; border-top: 1px solid #646464; border-bottom: 1px solid #646464; padding-top: 10px; padding-bottom: 14px;}
div.foot { font: 90% monospace; color: #787878; padding-top: 4px;}
</style>
</head>
<body>
<h2>Index of {{.Name}}</h2>
<div class="list">
<table summary="Directory Listing" cellpadding="0" cellspacing="0">
<thead><tr><th class="n">Name</th><th>Last Modified</th><th>Size (bytes)</th><th class="dl">Options</th></tr></thead>
<tbody>
{{range .Children_files}}
<tr><td class="n"><a href="/logs/{{.Path}}">{{.Name}}</a></td><td>{{.Mtime}}</td><td>{{.Size}}</td><td class="dl"><a href="/logs/{{.Path}}">Download</a></td></tr>
{{end}}
</tbody>
</table>
</div>
<div class="foot">{{.ServerUA}}</div>
</body>
</html>`

type fileInfo struct {
	Name  string
	Path  string
	Mtime string
	Size  string
}

// Manages directory listings
type dirlisting struct {
	Name           string
	Children_files []fileInfo
	ServerUA       string
}

// viewLogs serves log files from /var/log.
// Note: A future enhancement could show sessions from the sqlite database.
func viewLogs(w http.ResponseWriter, r *http.Request) {
	// Extract and clean the requested path
	urlpath := strings.TrimPrefix(r.URL.Path, "/logs/")

	// Build the full path using filepath.Join (which cleans the path)
	requestedPath := filepath.Join(varLogDirPath, urlpath)

	// Security: Validate that the resolved path is within /var/log
	// This prevents path traversal attacks (../, absolute paths, etc.)
	cleanPath := filepath.Clean(requestedPath)
	if !strings.HasPrefix(cleanPath, varLogDirPath) {
		log.Printf("viewLogs: path traversal attempt blocked: %s", r.URL.Path)
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	finfo, err := os.Stat(cleanPath)
	if err != nil {
		// Security: Use http.Error which automatically HTML-escapes the message
		// Only include the base filename in error message, not the full path
		http.Error(w, fmt.Sprintf("Failed to open %s: %s",
			filepath.Base(cleanPath), err.Error()),
			http.StatusNotFound)
		return
	}

	if !finfo.IsDir() {
		// NOSONAR: Path traversal is prevented by validation at lines 1029-1036
		// cleanPath is validated to start with varLogDirPath before reaching here
		http.ServeFile(w, r, cleanPath) //NOSONAR
		return
	}

	names, err := os.ReadDir(cleanPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read directory: %s", err.Error()),
			http.StatusInternalServerError)
		return
	}

	fi := make([]fileInfo, 0)
	for _, val := range names {
		if val.Name()[0] == '.' {
			continue
		} // Remove hidden files from listing

		info, err := val.Info()
		if err != nil {
			continue
		}
		if info.IsDir() {
			mtime := info.ModTime().Format("2006-Jan-02 15:04:05")
			sz := ""
			fi = append(fi, fileInfo{Name: val.Name() + "/", Path: urlpath + "/" + val.Name(), Mtime: mtime, Size: sz})
		} else {
			mtime := info.ModTime().Format("2006-Jan-02 15:04:05")
			sz := humanize.Comma(info.Size())
			fi = append(fi, fileInfo{Name: val.Name(), Path: urlpath + "/" + val.Name(), Mtime: mtime, Size: sz})
		}
	}

	tpl, err := template.New("tpl").Parse(dirlisting_tpl)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data := dirlisting{Name: r.URL.Path, ServerUA: "Stratux " + stratuxVersion + "/" + stratuxBuild,
		Children_files: fi}

	err = tpl.Execute(w, data)
	if err != nil {
		log.Printf("viewLogs() error: %s\n", err.Error())
	}

}

func connectMbTilesArchive(path string) (*sql.DB, map[string]string, error) {
	mbtileCacheLock.Lock()
	defer mbtileCacheLock.Unlock()
	if conn, ok := mbtileConnectionCache[path]; ok {
		if !conn.IsOutdated() {
			return conn.Conn, conn.Metadata, nil
		}
		log.Printf("Reloading MBTiles %s", path)
	}

	conn, err := sql.Open("sqlite3", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, nil, err
	}
	cacheEntry := NewMbTileConnectionCacheEntry(path, conn)
	cacheEntry.Metadata = readMbTilesMetadata(path, conn)
	if cacheEntry != nil {
		mbtileConnectionCache[path] = *cacheEntry
	}
	return conn, cacheEntry.Metadata, nil
}

func tileToDegree(z, x, y int) (lon, lat float64) {
	// osm-like schema:
	y = (1 << z) - y - 1
	n := math.Pi - 2.0*math.Pi*float64(y)/math.Exp2(float64(z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	lon = float64(x)/math.Exp2(float64(z))*360.0 - 180.0
	return lon, lat
}

func readMbTilesMetadata(fname string, db *sql.DB) map[string]string {
	rows, err := db.Query(`SELECT name, value FROM metadata 
		UNION SELECT 'minzoom', min(zoom_level) FROM tiles WHERE NOT EXISTS (SELECT * FROM metadata WHERE name='minzoom' and value is not null and value != '')
		UNION SELECT 'maxzoom', max(zoom_level) FROM tiles WHERE NOT EXISTS (SELECT * FROM metadata WHERE name='maxzoom' and value is not null and value != '')`)
	if err != nil {
		log.Printf("SQLite read error %s: %s", fname, err.Error())
		return nil
	}
	defer rows.Close()
	meta := make(map[string]string)
	for rows.Next() {
		var name, val string
		rows.Scan(&name, &val)
		if len(val) > 0 {
			meta[name] = val
		}
	}
	// determine extent of layer if not given.. Openlayers kinda needs this, or it can happen that it tries to do
	// a billion request do down-scale high-res pngs that aren't even there (i.e. all 404s)
	if _, ok := meta["bounds"]; !ok {
		maxZoomInt, err := strconv.ParseInt(meta["maxzoom"], 10, 32)
		if err != nil {
			log.Printf("SQLite metadata error for %s: invalid maxzoom value", fname)
			maxZoomInt = 0
		}
		rows, err = db.Query("SELECT min(tile_column), min(tile_row), max(tile_column), max(tile_row) FROM tiles WHERE zoom_level=?", maxZoomInt)
		if err != nil {
			log.Printf("SQLite read error %s: %s", fname, err.Error())
			return nil
		}
		rows.Next()
		var xmin, ymin, xmax, ymax int
		rows.Scan(&xmin, &ymin, &xmax, &ymax)
		lonmin, latmin := tileToDegree(int(maxZoomInt), xmin, ymin)
		lonmax, latmax := tileToDegree(int(maxZoomInt), xmax+1, ymax+1)
		meta["bounds"] = fmt.Sprintf("%f,%f,%f,%f", lonmin, latmin, lonmax, latmax)
	}

	// check if it is vectortiles and we have a style, then add the URL to metadata...
	if format, ok := meta["format"]; ok && format == "pbf" {
		_, file := filepath.Split(fname)
		if _, err := os.Stat(getMapdataStylesPath() + file + "/style.json"); err == nil {
			// We found a style!
			meta["stratux_style_url"] = mapdataStylesURLPath + file + "/style.json"
		}

	}
	return meta
}

// Scans mapdata dir for all .db and .mbtiles files and returns json representation of all metadata values
func handleTilesets(w http.ResponseWriter, _ *http.Request) {
	files, err := os.ReadDir(getMapdataPath())
	if err != nil {
		log.Printf("handleTilesets() error: %s\n", err.Error())
		http.Error(w, err.Error(), 500)
	}
	result := make(map[string]map[string]string, 0)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasSuffix(f.Name(), ".mbtiles") || strings.HasSuffix(f.Name(), ".db") {
			_, meta, err := connectMbTilesArchive(getMapdataPath() + f.Name())
			if err != nil {
				log.Printf("SQLite open "+f.Name()+" failed: %s", err.Error())
				continue
			}
			result[f.Name()] = meta
		}
	}
	resJson, err := json.Marshal(result)
	if err != nil {
		log.Printf("Error marshaling tilesets JSON: %s", err.Error())
	}
	w.Write(resJson)
}

func loadTile(fname string, z, x, y int) ([]byte, error) {
	db, meta, err := connectMbTilesArchive(getMapdataPath() + fname)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query("SELECT tile_data FROM tiles WHERE zoom_level=? AND tile_column=? AND tile_row=?", z, x, y)
	if err != nil {
		log.Printf("Failed to query mbtiles: %s", err.Error())
		return nil, nil
	}

	defer rows.Close()
	for rows.Next() {
		var res []byte
		rows.Scan(&res)
		// sometimes pbfs are gzipped...
		if format, ok := meta["format"]; ok && format == "pbf" && len(res) >= 2 && res[0] == 0x1f && res[1] == 0x8b {
			reader := bytes.NewReader(res)
			gzreader, _ := gzip.NewReader(reader)
			unzipped, err := io.ReadAll(gzreader)
			if err != nil {
				log.Printf("Failed to unzip gzipped PBF data")
				return nil, nil
			}
			res = unzipped
		}
		return res, nil
	}
	return nil, nil
}

func handleTile(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.RequestURI, "/")
	if len(parts) < 4 {
		return
	}
	idx := len(parts) - 1
	y, err := strconv.Atoi(strings.Split(parts[idx], ".")[0])
	if err != nil {
		http.Error(w, "Failed to parse y", 500)
		return
	}
	idx--
	x, err := strconv.Atoi(parts[idx])
	if err != nil {
		http.Error(w, "Failed to parse x", 500)
		return
	}
	idx--
	z, err := strconv.Atoi(parts[idx])
	if err != nil {
		http.Error(w, "Failed to parse z", 500)
		return
	}
	idx--
	file, err := url.QueryUnescape(parts[idx])
	if err != nil {
		http.Error(w, "Failed to parse file name", 500)
		return
	}
	tileData, err := loadTile(file, z, x, y)
	if err != nil {
		http.Error(w, err.Error(), 500)
	} else if tileData == nil {
		http.Error(w, "Tile not found", 404)
	} else {
		w.Write(tileData)
	}
}

func managementInterface() {
	weatherUpdate = NewUIBroadcaster()
	trafficUpdate = NewUIBroadcaster()
	radarUpdate = NewUIBroadcaster()
	situationUpdate = NewUIBroadcaster()
	weatherRawUpdate = NewUIBroadcaster()
	gdl90Update = NewUIBroadcaster()

	http.HandleFunc("/", defaultServer)
	//http.Handle("/logs/", http.StripPrefix("/logs/", http.FileServer(http.Dir("/var/log"))))
	http.Handle(mapdataStylesURLPath, http.StripPrefix(mapdataStylesURLPath, http.FileServer(http.Dir(getMapdataStylesPath()))))
	http.HandleFunc("/logs/", viewLogs)

	http.HandleFunc("/gdl90",
		func(w http.ResponseWriter, req *http.Request) {
			s := websocket.Server{
				Handler: websocket.Handler(handleGDL90WS)}
			s.ServeHTTP(w, req)
		})
	http.HandleFunc("/status",
		func(w http.ResponseWriter, req *http.Request) {
			s := websocket.Server{
				Handler: websocket.Handler(handleStatusWS)}
			s.ServeHTTP(w, req)
		})
	http.HandleFunc("/situation",
		func(w http.ResponseWriter, req *http.Request) {
			s := websocket.Server{
				Handler: websocket.Handler(handleSituationWS)}
			s.ServeHTTP(w, req)
		})
	http.HandleFunc("/weather",
		func(w http.ResponseWriter, req *http.Request) {
			s := websocket.Server{
				Handler: websocket.Handler(handleWeatherWS)}
			s.ServeHTTP(w, req)
		})
	http.HandleFunc("/traffic",
		func(w http.ResponseWriter, req *http.Request) {
			s := websocket.Server{
				Handler: websocket.Handler(handleTrafficWS)}
			s.ServeHTTP(w, req)
		})
	http.HandleFunc("/radar",
		func(w http.ResponseWriter, req *http.Request) {
			s := websocket.Server{
				Handler: websocket.Handler(handleRadarWS)}
			s.ServeHTTP(w, req)
		})

	http.HandleFunc("/jsonio",
		func(w http.ResponseWriter, req *http.Request) {
			s := websocket.Server{
				Handler: websocket.Handler(handleJsonIo)}
			s.ServeHTTP(w, req)
		})

	http.HandleFunc("/getStatus", handleStatusRequest)
	http.HandleFunc("/getSituation", handleSituationRequest)
	http.HandleFunc("/getTowers", handleTowersRequest)
	http.HandleFunc("/getSatellites", handleSatellitesRequest)
	http.HandleFunc("/getSettings", handleSettingsGetRequest)
	http.HandleFunc("/getRegion", handleRegionGet)
	http.HandleFunc("/setRegion", handleRegionSet)
	http.HandleFunc("/setSettings", handleSettingsSetRequest)
	http.HandleFunc("/restart", handleRestartRequest)
	http.HandleFunc("/shutdown", handleShutdownRequest)
	http.HandleFunc("/reboot", handleRebootRequest)
	http.HandleFunc("/getClients", handleClientsGetRequest)
	http.HandleFunc("/updateUpload", handleUpdatePostRequest)
	http.HandleFunc("/updatePong", handlePongUpdatePostRequest)
	http.HandleFunc("/roPartitionRebuild", handleroPartitionRebuild)
	http.HandleFunc("/develmodetoggle", handleDevelModeToggle)
	http.HandleFunc("/orientAHRS", handleOrientAHRS)
	http.HandleFunc("/calibrateAHRS", handleCalibrateAHRS)
	http.HandleFunc("/cageAHRS", handleCageAHRS)
	http.HandleFunc("/resetGMeter", handleResetGMeter)
	http.HandleFunc("/deletelogfile", handleDeleteLogFile)
	http.HandleFunc("/downloadlog", handleDownloadLogRequest)
	http.HandleFunc("/deleteahrslogfiles", handleDeleteAHRSLogFiles)
	http.HandleFunc("/downloadahrslogs", handleDownloadAHRSLogsRequest)
	http.HandleFunc("/downloaddb", handleDownloadDBRequest)
	http.HandleFunc("/tiles/tilesets", handleTilesets)
	http.HandleFunc("/tiles/", handleTile)

	addr := fmt.Sprintf(":%d", ManagementAddr)
	log.Printf("web configuration console on port %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("managementInterface ListenAndServe: %s\n", err.Error())
	}
}
