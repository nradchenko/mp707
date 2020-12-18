package main

import (
	"flag"
	"fmt"
	"github.com/nradchenko/mp707/onewire"
	"github.com/nradchenko/mp707/usb"
	"log"
)

func init() {
	if err := usb.InitLib(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	printHelp := flag.Bool("help", false, "show help")
	deviceId := flag.Uint("device", 0, "device ID")
	sensorId := flag.String("sensor", "", "sensor ROM ID")

	flag.Parse()

	if flag.Parsed() {
		if *printHelp {
			flag.PrintDefaults()
			return
		}

		devices, err := usb.Lookup()
		if err != nil {
			log.Fatal(err)
		}

		if len(devices) == 0 {
			log.Fatal("no devices found")
		}

		for _, device := range devices {
			if *deviceId > 0 {
				if device.GetId() != uint8(*deviceId) {
					log.Println("skipping device", device.GetId())
					continue
				}
			}

			log.Println("processing device", device.GetId())

			if *sensorId != "" {
				log.Println("processing sensor", *sensorId)

				if temperature, err := device.GetTemperature(onewire.StringToRom(*sensorId)); err != nil {
					log.Println(err, "(invalid ROM?)")
				} else {
					fmt.Println(*sensorId, temperature)
				}
			} else {
				if roms, err := device.GetSensors(); err != nil {
					log.Fatal(err)
				} else {
					for _, rom := range roms {
						log.Println("processing sensor", rom)

						if temperature, err := device.GetTemperature(rom); err != nil {
							log.Println(err)
						} else {
							fmt.Println(rom, temperature)
						}
					}
				}
			}

		}
	}
}
