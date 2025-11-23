package bluetooth

import (
	"encoding/binary"
	"errors"
)

// FTMS Indoor Bike Data flags
const (
	flagMoreData           uint16 = 1 << 0
	flagAverageSpeed       uint16 = 1 << 1
	flagInstCadence        uint16 = 1 << 2
	flagAvgCadence         uint16 = 1 << 3
	flagTotalDistance      uint16 = 1 << 4
	flagResistanceLevel    uint16 = 1 << 5
	flagInstPower          uint16 = 1 << 6
	flagAvgPower           uint16 = 1 << 7
	flagExpendedEnergy     uint16 = 1 << 8
	flagHeartRate          uint16 = 1 << 9
	flagMetabolicEquiv     uint16 = 1 << 10
	flagElapsedTime        uint16 = 1 << 11
	flagRemainingTime      uint16 = 1 << 12
)

// ParseIndoorBikeData parses FTMS Indoor Bike Data characteristic
func ParseIndoorBikeData(data []byte) (TrainerData, error) {
	if len(data) < 2 {
		return TrainerData{}, errors.New("data too short for flags")
	}

	flags := binary.LittleEndian.Uint16(data[0:2])
	offset := 2

	var result TrainerData

	// Instantaneous Speed is always present (uint16, 0.01 km/h resolution)
	if len(data) < offset+2 {
		return TrainerData{}, errors.New("data too short for speed")
	}
	// speed := float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.01
	offset += 2

	// Average Speed (optional)
	if flags&flagAverageSpeed != 0 {
		offset += 2
	}

	// Instantaneous Cadence (optional, uint16, 0.5 rpm resolution)
	if flags&flagInstCadence != 0 {
		if len(data) < offset+2 {
			return TrainerData{}, errors.New("data too short for cadence")
		}
		result.Cadence = float64(binary.LittleEndian.Uint16(data[offset:offset+2])) * 0.5
		offset += 2
	}

	// Average Cadence (optional)
	if flags&flagAvgCadence != 0 {
		offset += 2
	}

	// Total Distance (optional, uint24)
	if flags&flagTotalDistance != 0 {
		offset += 3
	}

	// Resistance Level (optional)
	if flags&flagResistanceLevel != 0 {
		offset += 2
	}

	// Instantaneous Power (optional, sint16, 1W resolution)
	if flags&flagInstPower != 0 {
		if len(data) < offset+2 {
			return TrainerData{}, errors.New("data too short for power")
		}
		result.Power = float64(int16(binary.LittleEndian.Uint16(data[offset : offset+2])))
		offset += 2
	}

	return result, nil
}
