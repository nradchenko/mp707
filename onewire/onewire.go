package onewire

import (
	"fmt"
	"strconv"
)

// 1-Wire ROM command set
// See DS1825 datasheet, ROM COMMANDS section (page 10-11)
// https://datasheets.maximintegrated.com/en/ds/DS1825.pdf
const (
	// Identify the ROM codes of all slave devices on the bus
	SearchRom = 0xf0
	// Read the slave’s 64-bit ROM code without using the Search ROM procedure (single slave device only)
	ReadRom = 0x33
	// Address a specific slave device on a multidrop or single-drop bus
	MatchRom = 0x55
	// Address all devices on the bus simultaneously without sending out any ROM code information
	SkipRom = 0xcc
	// Identify the ROM codes of all slave devices with a set alarm flag on the bus
	AlarmSearch = 0xec
)

// DS1825 function command set
// See DS1825 datasheet, table 5 (page 12)
// https://datasheets.maximintegrated.com/en/ds/DS1825.pdf
const (
	// Initiates temperature conversion
	ConvertT = 0x44
	// Reads the entire scratchpad including the CRC byte
	ReadScratchpad = 0xbe
	// Writes data into scratchpad bytes 2, 3, and 4 (TH, TL, and configuration registers)
	WriteScratchpad = 0x4e
	// Copies TH, TL, and configuration register data from the scratchpad to EEPROM
	CopyScratchpad = 0x48
	// Recalls TH, TL, and configuration register data from EEPROM to the scratchpad
	RecallE2 = 0xb8
	// Signals DS1825 power supply mode to the master
	ReadPowerSupply = 0xb4
)

// 1-Wire temperature sensors families
const (
	FamilyDS18S20 = 0x10
	FamilyDS1822  = 0x22
	FamilyDS18B20 = 0x28
)

// Rom represents 1-Wire device address
type Rom uint64

func (r Rom) String() string {
	return fmt.Sprintf("%x", uint64(r))
}

// StringToRom converts a human-readable 1-Wire device address representation to Rom type
func StringToRom(str string) Rom {
	rom, _ := strconv.ParseUint(str, 16, 64)
	return Rom(rom)
}

// Crc represents CRC checksum
type Crc uint64

// Crc8 calculates CRC8 checksum to check integrity of the data transferred over 1-Wire network
func Crc8(crc byte, d byte) Crc {
	r := crc

	for i := 0; i < 8; i++ {
		if ((r ^ (d >> i)) & 0x01) == 0x01 {
			r = ((r ^ 0x18) >> 1) | 0x80
		} else {
			r = (r >> 1) & 0x7f
		}
	}

	return Crc(r)
}

// CheckCrc8 verifies CRC8 checksum calculated by Crc8
func CheckCrc8(crc Crc) error {
	if crc != 0 {
		return CrcMismatchError{}
	}

	return nil
}

// GetFamilyByRom returns temperature sensor family
func GetFamilyByRom(rom Rom) byte {
	return byte(rom & 0xff)
}
