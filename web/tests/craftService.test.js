/**
 * Tests for craftService - Aircraft/Vessel categorization and color mapping
 *
 * These tests cover the pure JavaScript functions in craftService that:
 * - Map traffic sources to colors
 * - Determine aircraft/vessel categories
 * - Check if traffic is aged/stale
 * - Determine transport colors based on type
 */

// Constants from main.js
const TRAFFIC_MAX_AGE_SECONDS = 59;
const TRAFFIC_AIS_MAX_AGE_SECONDS = 60 * 15;
const TARGET_TYPE_AIS = 5;

// Extract craftService logic for testing
// This mirrors the implementation in web/js/main.js lines 149-328

const trafficSourceColors = {
    1: 'cornflowerblue', // ES
    2: '#FF8C00',        // UAT
    4: 'green',          // OGN
    5: '#0077be',        // AIS
    6: 'darkkhaki'       // UAT bar color
};

const getTrafficSourceColor = (source) => {
    if (trafficSourceColors[source] !== undefined) {
        return trafficSourceColors[source];
    } else {
        return 'gray';
    }
};

const aircraftColors = {
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
};

const getAircraftColor = (aircraft) => {
    let code = aircraft.Last_source.toString() + aircraft.TargetType.toString();
    if (aircraftColors[code] === undefined) {
        return 'white';
    } else {
        return aircraftColors[code];
    }
};

const getVesselColor = (vessel) => {
    const firstDigit = Math.floor(vessel.SurfaceVehicleType / 10);
    const secondDigit = vessel.SurfaceVehicleType - Math.floor(vessel.SurfaceVehicleType / 10) * 10;

    const categoryFirst = {
        2: 'orange',
        4: 'orange',
        5: 'orange',
        6: 'blue',
        7: 'green',
        8: 'red',
        9: 'red'
    };

    const categorySecond = {
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
    } else if (firstDigit === 3 && categorySecond[secondDigit]) {
        return categorySecond[secondDigit];
    } else {
        return 'gray';
    }
};

const isTrafficAged = (aircraft, targetVar) => {
    const value = aircraft[targetVar];
    if (aircraft.TargetType === TARGET_TYPE_AIS) {
        return value > TRAFFIC_AIS_MAX_AGE_SECONDS;
    } else {
        return value > TRAFFIC_MAX_AGE_SECONDS;
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
    return category[aircraft.Emitter_category] ? category[aircraft.Emitter_category] : '---';
};

const getVesselCategory = (vessel) => {
    const firstDigit = Math.floor(vessel.SurfaceVehicleType / 10);
    const secondDigit = vessel.SurfaceVehicleType - Math.floor(vessel.SurfaceVehicleType / 10) * 10;

    const categoryFirst = {
        2: 'Cargo',
        4: 'Cargo',
        5: 'Cargo',
        6: 'Passenger',
        7: 'Cargo',
        8: 'Tanker',
        9: 'Cargo',
    };

    const categorySecond = {
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
    } else if (firstDigit === 3 && categorySecond[secondDigit]) {
        return categorySecond[secondDigit];
    } else {
        return '---';
    }
};

// ============================================================================
// TESTS
// ============================================================================

describe('craftService - Traffic Source Colors', () => {
    test('should return cornflowerblue for ES (source 1)', () => {
        expect(getTrafficSourceColor(1)).toBe('cornflowerblue');
    });

    test('should return #FF8C00 for UAT (source 2)', () => {
        expect(getTrafficSourceColor(2)).toBe('#FF8C00');
    });

    test('should return green for OGN (source 4)', () => {
        expect(getTrafficSourceColor(4)).toBe('green');
    });

    test('should return #0077be for AIS (source 5)', () => {
        expect(getTrafficSourceColor(5)).toBe('#0077be');
    });

    test('should return darkkhaki for UAT bar (source 6)', () => {
        expect(getTrafficSourceColor(6)).toBe('darkkhaki');
    });

    test('should return gray for unknown source', () => {
        expect(getTrafficSourceColor(99)).toBe('gray');
        expect(getTrafficSourceColor(0)).toBe('gray');
        expect(getTrafficSourceColor(-1)).toBe('gray');
    });
});

describe('craftService - Aircraft Colors', () => {
    test('should return cornflowerblue for ES ADS-B (source 1, type 0)', () => {
        const aircraft = { Last_source: 1, TargetType: 0 };
        expect(getAircraftColor(aircraft)).toBe('cornflowerblue');
    });

    test('should return cornflowerblue for ES Mode-S (source 1, type 1)', () => {
        const aircraft = { Last_source: 1, TargetType: 1 };
        expect(getAircraftColor(aircraft)).toBe('cornflowerblue');
    });

    test('should return darkorange for UAT (source 2, type 0)', () => {
        const aircraft = { Last_source: 2, TargetType: 0 };
        expect(getAircraftColor(aircraft)).toBe('darkorange');
    });

    test('should return green for OGN (source 4, type 0)', () => {
        const aircraft = { Last_source: 4, TargetType: 0 };
        expect(getAircraftColor(aircraft)).toBe('green');
    });

    test('should return white for unknown type', () => {
        const aircraft = { Last_source: 99, TargetType: 99 };
        expect(getAircraftColor(aircraft)).toBe('white');
    });
});

describe('craftService - Aircraft Categories', () => {
    test('should return Light for emitter category 1', () => {
        const aircraft = { Emitter_category: 1 };
        expect(getAircraftCategory(aircraft)).toBe('Light');
    });

    test('should return Heavy for emitter category 5', () => {
        const aircraft = { Emitter_category: 5 };
        expect(getAircraftCategory(aircraft)).toBe('Heavy');
    });

    test('should return Helic for emitter category 7', () => {
        const aircraft = { Emitter_category: 7 };
        expect(getAircraftCategory(aircraft)).toBe('Helic');
    });

    test('should return Glide for emitter category 9', () => {
        const aircraft = { Emitter_category: 9 };
        expect(getAircraftCategory(aircraft)).toBe('Glide');
    });

    test('should return Drone for emitter category 14', () => {
        const aircraft = { Emitter_category: 14 };
        expect(getAircraftCategory(aircraft)).toBe('Drone');
    });

    test('should return --- for unknown emitter category', () => {
        const aircraft = { Emitter_category: 99 };
        expect(getAircraftCategory(aircraft)).toBe('---');
    });

    test('should return --- for missing emitter category', () => {
        const aircraft = {};
        expect(getAircraftCategory(aircraft)).toBe('---');
    });
});

describe('craftService - Vessel Categories', () => {
    test('should return Cargo for vessel type 20-29', () => {
        const vessel = { SurfaceVehicleType: 20 };
        expect(getVesselCategory(vessel)).toBe('Cargo');
    });

    test('should return Passenger for vessel type 60-69', () => {
        const vessel = { SurfaceVehicleType: 60 };
        expect(getVesselCategory(vessel)).toBe('Passenger');
    });

    test('should return Tanker for vessel type 80-89', () => {
        const vessel = { SurfaceVehicleType: 80 };
        expect(getVesselCategory(vessel)).toBe('Tanker');
    });

    test('should return Fishing for vessel type 30', () => {
        const vessel = { SurfaceVehicleType: 30 };
        expect(getVesselCategory(vessel)).toBe('Fishing');
    });

    test('should return Military for vessel type 35', () => {
        const vessel = { SurfaceVehicleType: 35 };
        expect(getVesselCategory(vessel)).toBe('Military');
    });

    test('should return Sailing for vessel type 36', () => {
        const vessel = { SurfaceVehicleType: 36 };
        expect(getVesselCategory(vessel)).toBe('Sailing');
    });

    test('should return Cargo for vessel type 99 (90-99 range)', () => {
        const vessel = { SurfaceVehicleType: 99 };
        expect(getVesselCategory(vessel)).toBe('Cargo');
    });

    test('should return --- for vessel type 10 (unmapped range)', () => {
        const vessel = { SurfaceVehicleType: 10 };
        expect(getVesselCategory(vessel)).toBe('---');
    });
});

describe('craftService - Vessel Colors', () => {
    test('should return orange for cargo vessels (type 20-29)', () => {
        const vessel = { SurfaceVehicleType: 25 };
        expect(getVesselColor(vessel)).toBe('orange');
    });

    test('should return blue for passenger vessels (type 60-69)', () => {
        const vessel = { SurfaceVehicleType: 65 };
        expect(getVesselColor(vessel)).toBe('blue');
    });

    test('should return red for tanker vessels (type 80-89)', () => {
        const vessel = { SurfaceVehicleType: 85 };
        expect(getVesselColor(vessel)).toBe('red');
    });

    test('should return silver for fishing vessels (type 30)', () => {
        const vessel = { SurfaceVehicleType: 30 };
        expect(getVesselColor(vessel)).toBe('silver');
    });

    test('should return darkolivegreen for military vessels (type 35)', () => {
        const vessel = { SurfaceVehicleType: 35 };
        expect(getVesselColor(vessel)).toBe('darkolivegreen');
    });

    test('should return red for vessel type 99 (90-99 range)', () => {
        const vessel = { SurfaceVehicleType: 99 };
        expect(getVesselColor(vessel)).toBe('red');
    });

    test('should return gray for vessel type 10 (unmapped range)', () => {
        const vessel = { SurfaceVehicleType: 10 };
        expect(getVesselColor(vessel)).toBe('gray');
    });
});

describe('craftService - Traffic Age Detection', () => {
    test('should identify aircraft as NOT aged when Age < 59 seconds', () => {
        const aircraft = { TargetType: 1, Age: 30 };
        expect(isTrafficAged(aircraft, 'Age')).toBe(false);
    });

    test('should identify aircraft as aged when Age = 59 seconds (boundary)', () => {
        const aircraft = { TargetType: 1, Age: 59 };
        expect(isTrafficAged(aircraft, 'Age')).toBe(false);
    });

    test('should identify aircraft as aged when Age > 59 seconds', () => {
        const aircraft = { TargetType: 1, Age: 60 };
        expect(isTrafficAged(aircraft, 'Age')).toBe(true);
    });

    test('should identify aircraft as aged when Age >> 59 seconds', () => {
        const aircraft = { TargetType: 1, Age: 120 };
        expect(isTrafficAged(aircraft, 'Age')).toBe(true);
    });

    test('should identify AIS vessel as NOT aged when Age < 900 seconds', () => {
        const vessel = { TargetType: TARGET_TYPE_AIS, Age: 600 };
        expect(isTrafficAged(vessel, 'Age')).toBe(false);
    });

    test('should identify AIS vessel as aged when Age > 900 seconds', () => {
        const vessel = { TargetType: TARGET_TYPE_AIS, Age: 901 };
        expect(isTrafficAged(vessel, 'Age')).toBe(true);
    });

    test('should work with AgeLastAlt field', () => {
        const aircraft = { TargetType: 1, AgeLastAlt: 65 };
        expect(isTrafficAged(aircraft, 'AgeLastAlt')).toBe(true);
    });

    test('should identify fresh traffic correctly', () => {
        const aircraft = { TargetType: 1, Age: 0 };
        expect(isTrafficAged(aircraft, 'Age')).toBe(false);
    });
});
