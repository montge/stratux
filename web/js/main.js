// application constants
const URL_HOST_BASE           = window.location.hostname + (window.location.port ? ':' + window.location.port : '');
const URL_HOST_PROTOCOL       = window.location.protocol + "//";

const URL_AHRS_CAGE           = URL_HOST_PROTOCOL + URL_HOST_BASE + "/cageAHRS";
const URL_AHRS_CAL            = URL_HOST_PROTOCOL + URL_HOST_BASE + "/calibrateAHRS";
const URL_AHRS_ORIENT         = URL_HOST_PROTOCOL + URL_HOST_BASE + "/orientAHRS";
const URL_DELETEAHRSLOGFILES  = URL_HOST_PROTOCOL + URL_HOST_BASE + "/deleteahrslogfiles";
const URL_DELETELOGFILE       = URL_HOST_PROTOCOL + URL_HOST_BASE + "/deletelogfile";
const URL_DEV_TOGGLE_GET      = URL_HOST_PROTOCOL + URL_HOST_BASE + "/develmodetoggle";
const URL_DOWNLOADAHRSLOGFILES = URL_HOST_PROTOCOL + URL_HOST_BASE + "/downloadahrslogs";
const URL_DOWNLOADDB          = URL_HOST_PROTOCOL + URL_HOST_BASE + "/downloaddb";
const URL_DOWNLOADLOGFILE     = URL_HOST_PROTOCOL + URL_HOST_BASE + "/downloadlog";
const URL_GMETER_RESET        = URL_HOST_PROTOCOL + URL_HOST_BASE + "/resetGMeter";
const URL_REBOOT              = URL_HOST_PROTOCOL + URL_HOST_BASE + "/reboot";
const URL_RESTARTAPP          = URL_HOST_PROTOCOL + URL_HOST_BASE + "/restart";
const URL_SATELLITES_GET      = URL_HOST_PROTOCOL + URL_HOST_BASE + "/getSatellites";
const URL_SETTINGS_GET        = URL_HOST_PROTOCOL + URL_HOST_BASE + "/getSettings";
const URL_SETTINGS_SET        = URL_HOST_PROTOCOL + URL_HOST_BASE + "/setSettings";
const URL_SHUTDOWN            = URL_HOST_PROTOCOL + URL_HOST_BASE + "/shutdown";
const URL_STATUS_GET          = URL_HOST_PROTOCOL + URL_HOST_BASE + "/getStatus";
const URL_REGION_GET          = URL_HOST_PROTOCOL + URL_HOST_BASE + "/getRegion";
const URL_REGION_SET          = URL_HOST_PROTOCOL + URL_HOST_BASE + "/setRegion";
const URL_TOWERS_GET          = URL_HOST_PROTOCOL + URL_HOST_BASE + "/getTowers";
const URL_UPDATE_UPLOAD       = URL_HOST_PROTOCOL + URL_HOST_BASE + "/updateUpload";
const URL_UPDATE_PONG         = URL_HOST_PROTOCOL + URL_HOST_BASE + "/updatePong";
const URL_GET_SITUATION       = URL_HOST_PROTOCOL + URL_HOST_BASE + "/getSituation";
const URL_GET_TILESETS        = URL_HOST_PROTOCOL + URL_HOST_BASE + "/tiles/tilesets";
const URL_GET_TILE            = URL_HOST_PROTOCOL + URL_HOST_BASE + "/tiles";
const URL_GET_STYLE           = URL_HOST_PROTOCOL + URL_HOST_BASE + "/mapdata/styles"


const URL_DEVELOPER_WS        = "ws://" + URL_HOST_BASE + "/developer";
const URL_GPS_WS              = "ws://" + URL_HOST_BASE + "/situation";
const URL_STATUS_WS           = "ws://" + URL_HOST_BASE + "/status";
const URL_TRAFFIC_WS          = "ws://" + URL_HOST_BASE + "/traffic";
const URL_WEATHER_WS          = "ws://" + URL_HOST_BASE + "/weather";
const URL_RADAR_WS            = "ws://" + URL_HOST_BASE + "/radar";

// define the module with dependency on mobile-angular-ui
//var app = angular.module('stratux', ['ngRoute', 'mobile-angular-ui', 'mobile-angular-ui.gestures', 'appControllers']);
var app = angular.module('stratux', ['ui.router', 'mobile-angular-ui', 'mobile-angular-ui.gestures', 'appControllers']);
var appControllers = angular.module('appControllers', []);
//cutoff value to remove targets out of the list, keep in sync with the value in traffic.go for cleanUpOldEntries, keep it just below cutoff value in traffic.go
let TRAFFIC_MAX_AGE_SECONDS = 59;
let TRAFFIC_AIS_MAX_AGE_SECONDS = 60*15;
let TARGET_TYPE_AIS = 5;

app.config(function ($stateProvider, $urlRouterProvider) {
	$stateProvider
		.state('home', {
			url: '/',
			templateUrl: 'plates/status.html',
			controller: 'StatusCtrl',
			reloadOnSearch: false
		})
		.state('towers', {
			url: '/towers',
			templateUrl: 'plates/towers.html',
			controller: 'TowersCtrl',
			reloadOnSearch: false
		})
		.state('weather', {
			url: '/weather',
			templateUrl: 'plates/weather.html',
			controller: 'WeatherCtrl',
			reloadOnSearch: false
		})
		.state('traffic', {
			url: '/traffic',
			templateUrl: 'plates/traffic.html',
			controller: 'TrafficCtrl',
			reloadOnSearch: false
		})
		.state('gps', {
			url: '/gps',
			templateUrl: 'plates/gps.html',
			controller: 'GPSCtrl',
			reloadOnSearch: false
		})
		.state('logs', {
			url: '/logs',
			templateUrl: 'plates/logs.html',
			controller: 'LogsCtrl',
			reloadOnSearch: false
		})
		.state('settings', {
			url: '/settings',
			templateUrl: 'plates/settings.html',
			controller: 'SettingsCtrl',
			reloadOnSearch: false
		})
		.state('radar', {
			url: '/radar',
			templateUrl: 'plates/radar.html',
			controller: 'RadarCtrl',
			reloadOnSearch: false
		})
		.state('map', {
			url: '/map',
			templateUrl: 'plates/map.html',
			controller: 'MapCtrl',
			reloadOnSearch: false
		})
        .state('developer', {
			url: '/developer',
			templateUrl: 'plates/developer.html',
			controller: 'DeveloperCtrl',
			reloadOnSearch: false
		});
	$urlRouterProvider.otherwise('/');
});


app.run(function ($transform) {
	window.$transform = $transform;
});

// For this app we have a MainController for whatever and individual controllers for each page
app.controller('MainCtrl', function ($scope, $http) {
	// any logic global logic
    $http.get(URL_SETTINGS_GET)
    .then(function(response) {
			var settings = angular.fromJson(response.data);
            $scope.DeveloperMode = settings.DeveloperMode;
            $scope.UAT_Enabled = settings.UAT_Enabled;
            $scope.Ping_Enabled = settings.Ping_Enabled;
            $scope.Pong_Enabled = settings.Pong_Enabled;

            // Update theme
            $scope.updateTheme(settings.DarkMode);
    }, function(response) {
        //Second function handles error
    });	

    $scope.updateTheme = function(darkMode) {
        if(darkMode != $scope.DarkMode) {
            $scope.DarkMode = darkMode;

            if($scope.DarkMode) {
                document.getElementById('themeStylesheet').href = 'css/dark-mode.css';
            } else {
                document.getElementById('themeStylesheet').href = '';
            }
        }
    };
})
.service('craftService',function(){ 
	let trafficSourceColors = {
		1: 'cornflowerblue', // ES
		2: '#FF8C00',      // UAT
		4: 'green',          // OGN
		5: '#0077be',         // AIS
		6: 'darkkhaki'     // UAT bar color
	}

	const getTrafficSourceColor = (source) => {
		if (trafficSourceColors[source] !== undefined) {
			return trafficSourceColors[source];
		} else {
			return 'gray';
		}
	}

	// THis ensures that the colors used in traffic.js and map.js for the vessels are the same
	let aircraftColors = {

		10: 'cornflowerblue',
		11: 'cornflowerblue',
		12: 'skyblue',
		13: 'skyblue',
		14: 'skyblue',

		20: 'darkorange',
		21: 'darkorange',
		22: 'orange',
		23: 'orange',
		24: 'orange',

		40: 'green',
		41: 'green',
		42: 'greenyellow',
		43: 'greenyellow',
		44: 'greenyellow'
	}

	const getAircraftColor = (aircraft) => {
		let code = aircraft.Last_source.toString()+aircraft.TargetType.toString();			
		if (aircraftColors[code] === undefined) {
			return 'white';
		} else {
			return aircraftColors[code];
		}
	};

	const getVesselColor = (vessel) => {
		// https://www.navcen.uscg.gov/?pageName=AISMessagesAStatic
		firstDigit = Math.floor(vessel.SurfaceVehicleType / 10)
		secondDigit = vessel.SurfaceVehicleType - Math.floor(vessel.SurfaceVehicleType / 10)*10;

		const categoryFirst= {
			2: 'orange',
			4: 'orange',
			5: 'orange',
			6: 'blue',
			7: 'green',
			8: 'red',
			9: 'red'
		};		
		const categorySecond= {
			0: 'silver',
			1: 'cyan',
			2: 'darkblue',
			3: 'LightSkyBlue',
			4: 'LightSkyBlue',
			5: 'darkolivegreen',
			6: 'maroon',
			7: 'purple'
		};		

		if (categoryFirst[firstDigit]) {
			return categoryFirst[firstDigit];
		} else if (firstDigit===3 && categorySecond[secondDigit]) {
			return categorySecond[secondDigit];
		} else {
			return 'gray';			
		}
	};

	const isTrafficAged = (aircraft, targetVar ) => {
		const value = aircraft[targetVar];
		if (aircraft.TargetType === TARGET_TYPE_AIS) {
			return value > TRAFFIC_AIS_MAX_AGE_SECONDS;
		} else { 
			return value > TRAFFIC_MAX_AGE_SECONDS;
		}
	};

	const getVesselCategory = (vessel) => {
		// https://www.navcen.uscg.gov/?pageName=AISMessagesAStatic
		firstDigit = Math.floor(vessel.SurfaceVehicleType / 10)
		secondDigit = vessel.SurfaceVehicleType - Math.floor(vessel.SurfaceVehicleType / 10)*10;

		const categoryFirst= {
			2: 'Cargo',
			4: 'Cargo',
			5: 'Cargo',
			6: 'Passenger',
			7: 'Cargo',
			8: 'Tanker',
			9: 'Cargo',
		};		
		const categorySecond= {
			0: 'Fishing',
			1: 'Tugs',
			2: 'Tugs',
			3: 'Dredging',
			4: 'Diving',
			5: 'Military',
			6: 'Sailing',
			7: 'Pleasure',
		};		

		if (categoryFirst[firstDigit]) {
			return categoryFirst[firstDigit];
		} else if (firstDigit===3 && categorySecond[secondDigit]) {
			return categorySecond[secondDigit];
		} else {
			return '---';			
		}
	};

	const getAircraftCategory = (aircraft) => {
		const category = {
			1: 'Light',
			2: 'Small',
			3: 'Large',
			4: 'VLarge',
			5: 'Heavy',
			6: 'Fight',
			7: 'Helic',
			9: 'Glide',
			10: 'Ballo',
			11: 'Parac',
			12: 'Ultrl',
			14: 'Drone',
			15: 'Space',
			16: 'VLarge',
			17: 'Vehic',
			18: 'Vehic',
			19: 'Obstc'
		};		
		return category[aircraft.Emitter_category]?category[aircraft.Emitter_category]:'---';
	};

	return {
		getCategory: (craft) => {
			if (craft.TargetType === TARGET_TYPE_AIS) {
				return getVesselCategory(craft);
			} else {
				return getAircraftCategory(craft);
			}
		},

		getTrafficSourceColor: (source) => {
			return getTrafficSourceColor(source);
		},

		isTrafficAged: (craft) => {
			return isTrafficAged(craft, 'Age');
		},

		isTrafficAged2: (craft, targetVar) => {
			return isTrafficAged(craft, targetVar);
		},

		getTransportColor: (craft) => {
			if (craft.TargetType === TARGET_TYPE_AIS) {
				return getVesselColor(craft);
			} else {
				return getAircraftColor(craft);
			}
		}
	
	};

});
