/**
 * Tests for traffic.js utility functions
 *
 * These tests cover pure JavaScript utility functions from traffic.js:
 * - UTC time string formatting
 * - DMS (Degrees Minutes Seconds) coordinate formatting
 * - Aircraft comparison logic
 */

// Extract utility functions from traffic.js for testing

/**
 * Convert epoch timestamp to UTC time string (HH:MM:SSZ format)
 * From traffic.js lines 17-29
 */
const utcTimeString = (epoc) => {
    let time = "";
    let val;
    const d = new Date(epoc);
    val = d.getUTCHours();
    time += (val < 10 ? "0" + val : "" + val);
    val = d.getUTCMinutes();
    time += ":" + (val < 10 ? "0" + val : "" + val);
    val = d.getUTCSeconds();
    time += ":" + (val < 10 ? "0" + val : "" + val);
    time += "Z";
    return time;
};

/**
 * Convert decimal degrees to DMS (Degrees Minutes) format
 * From traffic.js lines 43-53
 */
const dmsString = (val) => {
    let deg;
    let min;
    deg = 0 | val;
    min = 0 | (val < 0 ? val = -val : val) % 1 * 60;

    return [deg*deg < 100 ? "0" + deg : deg,
            '° ',
            min < 10 ? "0" + min : min,
            "' "].join('');
};

/**
 * Check if two aircraft are the same based on address and type
 * From traffic.js lines 112-120
 */
const isSameAircraft = (addr1, addrType1, addr2, addrType2) => {
    if (addr1 != addr2)
        return false;
    // Both aircraft have the same address and it is either an ICAO address for both,
    // or a non-icao address for both.
    // 1 = non-icao, everything else = icao
    if ((addrType1 == 1 && addrType2 == 1) || (addrType1 != 1 && addrType2 != 1))
        return true;
};

// ============================================================================
// TESTS
// ============================================================================

describe('traffic.js - UTC Time String Formatting', () => {
    test('should format midnight UTC correctly', () => {
        const midnight = Date.parse('2025-10-13T00:00:00Z');
        expect(utcTimeString(midnight)).toBe('00:00:00Z');
    });

    test('should format noon UTC correctly', () => {
        const noon = Date.parse('2025-10-13T12:00:00Z');
        expect(utcTimeString(noon)).toBe('12:00:00Z');
    });

    test('should format single digit hours with leading zero', () => {
        const morning = Date.parse('2025-10-13T09:15:30Z');
        expect(utcTimeString(morning)).toBe('09:15:30Z');
    });

    test('should format single digit minutes with leading zero', () => {
        const time = Date.parse('2025-10-13T12:05:30Z');
        expect(utcTimeString(time)).toBe('12:05:30Z');
    });

    test('should format single digit seconds with leading zero', () => {
        const time = Date.parse('2025-10-13T12:15:05Z');
        expect(utcTimeString(time)).toBe('12:15:05Z');
    });

    test('should format time with all single digits correctly', () => {
        const time = Date.parse('2025-10-13T01:02:03Z');
        expect(utcTimeString(time)).toBe('01:02:03Z');
    });

    test('should format time with all double digits correctly', () => {
        const time = Date.parse('2025-10-13T23:59:59Z');
        expect(utcTimeString(time)).toBe('23:59:59Z');
    });

    test('should always end with Z', () => {
        const time = Date.parse('2025-10-13T12:00:00Z');
        const result = utcTimeString(time);
        expect(result.endsWith('Z')).toBe(true);
    });
});

describe('traffic.js - DMS Coordinate Formatting', () => {
    test('should format zero degrees correctly', () => {
        expect(dmsString(0)).toBe('00° 00\' ');
    });

    test('should format positive single digit degrees', () => {
        expect(dmsString(5.5)).toBe('05° 30\' ');
    });

    test('should format positive double digit degrees', () => {
        expect(dmsString(47.5)).toBe('47° 30\' ');
    });

    test('should format positive triple digit degrees', () => {
        expect(dmsString(122.5)).toBe('122° 30\' ');
    });

    test('should handle negative coordinates (latitude)', () => {
        // Negative coordinates should be converted to positive in DMS
        const result = dmsString(-47.5);
        expect(result).toContain('47°');
        expect(result).toContain('30\'');
    });

    test('should format Seattle latitude correctly (47.45N)', () => {
        const result = dmsString(47.45);
        expect(result).toBe('47° 27\' ');
    });

    test('should format Seattle longitude correctly (122.31W)', () => {
        const result = dmsString(-122.31);
        expect(result).toContain('122°');
        expect(result).toContain('18\'');
    });

    test('should format coordinates with no minutes', () => {
        expect(dmsString(47.0)).toBe('47° 00\' ');
    });

    test('should format coordinates with 59 minutes', () => {
        expect(dmsString(47.9833)).toBe('47° 58\' '); // 0.9833 * 60 ≈ 59
    });

    test('should always include degree and minute symbols', () => {
        const result = dmsString(45.5);
        expect(result).toContain('°');
        expect(result).toContain('\'');
    });

    test('should pad single digit degrees with leading zero (< 10)', () => {
        expect(dmsString(5.25)).toBe('05° 15\' ');
    });

    test('should pad single digit minutes with leading zero', () => {
        expect(dmsString(47.0833)).toBe('47° 04\' '); // 0.0833 * 60 ≈ 5
    });
});

describe('traffic.js - Aircraft Comparison', () => {
    test('should return false for different addresses', () => {
        expect(isSameAircraft(0xA12345, 0, 0xAC82EC, 0)).toBe(false);
    });

    test('should return true for same address with ICAO type (0)', () => {
        expect(isSameAircraft(0xA12345, 0, 0xA12345, 0)).toBe(true);
    });

    test('should return true for same address with non-ICAO type (1)', () => {
        expect(isSameAircraft(0xA12345, 1, 0xA12345, 1)).toBe(true);
    });

    test('should return false for same address but mixed types (ICAO vs non-ICAO)', () => {
        expect(isSameAircraft(0xA12345, 0, 0xA12345, 1)).toBe(undefined);
        expect(isSameAircraft(0xA12345, 1, 0xA12345, 0)).toBe(undefined);
    });

    test('should treat type 0 and type 2 as both ICAO (not type 1)', () => {
        expect(isSameAircraft(0xA12345, 0, 0xA12345, 2)).toBe(true);
    });

    test('should treat type 3 and type 4 as both ICAO', () => {
        expect(isSameAircraft(0xA12345, 3, 0xA12345, 4)).toBe(true);
    });

    test('should handle address 0 (invalid/unknown)', () => {
        expect(isSameAircraft(0, 0, 0, 0)).toBe(true);
    });

    test('should handle very large ICAO addresses (24-bit)', () => {
        expect(isSameAircraft(0xFFFFFF, 0, 0xFFFFFF, 0)).toBe(true);
    });
});

describe('traffic.js - Edge Cases', () => {
    describe('utcTimeString edge cases', () => {
        test('should handle epoch 0 (1970-01-01)', () => {
            const result = utcTimeString(0);
            expect(result).toBe('00:00:00Z');
        });

        test('should handle current time', () => {
            const now = Date.now();
            const result = utcTimeString(now);
            // Should be a valid time string format
            expect(result).toMatch(/^\d{2}:\d{2}:\d{2}Z$/);
        });
    });

    describe('dmsString edge cases', () => {
        test('should handle very small decimal values', () => {
            const result = dmsString(0.01);
            expect(result).toBe('00° 00\' ');
        });

        test('should handle 180 degrees (max longitude)', () => {
            const result = dmsString(180);
            expect(result).toBe('180° 00\' ');
        });

        test('should handle -180 degrees', () => {
            const result = dmsString(-180);
            expect(result).toContain('180°');
        });

        test('should handle 90 degrees (max latitude)', () => {
            const result = dmsString(90);
            expect(result).toBe('90° 00\' ');
        });
    });

    describe('isSameAircraft edge cases', () => {
        test('should handle null address type gracefully', () => {
            // Type 0 is treated as ICAO
            expect(isSameAircraft(0xA12345, 0, 0xA12345, 0)).toBe(true);
        });

        test('should return undefined for mixed ICAO/non-ICAO comparison', () => {
            const result = isSameAircraft(0xA12345, 0, 0xA12345, 1);
            expect(result).toBe(undefined);
        });
    });
});
