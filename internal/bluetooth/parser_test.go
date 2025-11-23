package bluetooth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIndoorBikeData_PowerAndCadence(t *testing.T) {
	// Flags: bit 2 (cadence) and bit 6 (power) set = 0x44 = 0b01000100
	// Data format: flags(2) + speed(2) + cadence(2) + power(2)
	// Speed always present (not optional)
	data := []byte{
		0x44, 0x00, // Flags: cadence + power present
		0xE8, 0x03, // Speed: 1000 (10.00 km/h, 0.01 resolution)
		0xB4, 0x00, // Cadence: 180 (90 rpm, 0.5 resolution)
		0xC8, 0x00, // Power: 200 watts
	}

	result, err := ParseIndoorBikeData(data)

	assert.NoError(t, err)
	assert.InDelta(t, 200.0, result.Power, 0.1)
	assert.InDelta(t, 90.0, result.Cadence, 0.1)
}

func TestParseIndoorBikeData_PowerOnly(t *testing.T) {
	// Flags: only bit 6 (power) set = 0x40
	data := []byte{
		0x40, 0x00, // Flags: power present only
		0xE8, 0x03, // Speed: 1000
		0x96, 0x00, // Power: 150 watts
	}

	result, err := ParseIndoorBikeData(data)

	assert.NoError(t, err)
	assert.InDelta(t, 150.0, result.Power, 0.1)
	assert.InDelta(t, 0.0, result.Cadence, 0.1)
}

func TestParseIndoorBikeData_TooShort(t *testing.T) {
	data := []byte{0x44} // Too short

	_, err := ParseIndoorBikeData(data)

	assert.Error(t, err)
}

func TestEncodeRequestControl(t *testing.T) {
	data := EncodeRequestControl()
	assert.Equal(t, []byte{0x00}, data)
}

func TestEncodeSetTargetResistance(t *testing.T) {
	// 50% resistance = 100 (0.1% resolution, so 50% = 500, but range is 0-200)
	// Actually: level 0-100 maps to 0-200 in protocol
	data := EncodeSetTargetResistance(50)
	assert.Equal(t, []byte{0x04, 100}, data)
}

func TestEncodeSetTargetPower(t *testing.T) {
	// 200 watts
	data := EncodeSetTargetPower(200)
	assert.Equal(t, []byte{0x05, 0xC8, 0x00}, data)
}
