package usb

import (
	"testing"
)

func TestUSB(t *testing.T) {
	if err := InitLib(); err != nil {
		t.Error(err)
	}

	defer DesposeLib()

	devices, err := Lookup()
	if err != nil {
		t.Errorf("could not enumerate devices: %s", err)
	}

	if len(devices) == 0 {
		t.Error("no devices present, can not test anything")
	}

	for _, device := range devices {
		t.Logf("processing device #%d", device.id)

		if roms, err := device.GetSensors(); err != nil {
			t.Errorf("failed to enumerate sensors: %s", err)
		} else {
			for _, rom := range roms {
				t.Log("processing sensor", rom)

				if temperature, err := device.GetTemperature(rom); err != nil {
					t.Errorf("failed to read temperature: %s", err)
				} else {
					t.Log("got temperature reading:", temperature)
				}
			}
		}
	}
}
