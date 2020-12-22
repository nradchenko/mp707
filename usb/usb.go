/*

Low level MP707 USB thermometer board functions

Based on protocol implementation in:
 https://github.com/joe-skb7/bmcontrol
 https://silines.ru/software/RODOS/RODOS-5_6/RODOS5_6.tar.gz

*/

package usb

/*
 #cgo pkg-config: libusb-1.0
 #include <libusb-1.0/libusb.h>

 libusb_device *get_dev(libusb_device **list, int offset) {
    return list[offset];
 }
*/
import "C"

import (
	"errors"
	"fmt"
	"github.com/nradchenko/mp707/onewire"
	"time"
)

// VID and PID of Rodos-5Z USB thermometer device
const (
	VendorId  = 0x20a0
	ProductId = 0x4173
)

// General settings
const (
	// Timeout for USB read/write operations
	Timeout = 5000
	// Number of retries for some operations
	Retries = 3
	// Number of ROM ports
	RomPorts = 64
	// Size of input and output buffers
	CmdSize = 8
)

// Common delays between USB operations
const (
	// Base delay
	DELAY_BASE = 1 * time.Millisecond
	// Single delay
	DELAY_SINGLE = DELAY_BASE
	// Double delay (4 bytes)
	DELAY_DOUBLE = 2 * DELAY_SINGLE
)

// libusb constants
const (
	LIBUSB_ENDPOINT_IN      = C.LIBUSB_ENDPOINT_IN
	LIBUSB_ENDPOINT_OUT     = C.LIBUSB_ENDPOINT_OUT
	LIBUSB_RECIPIENT_DEVICE = C.LIBUSB_RECIPIENT_DEVICE
	LIBUSB_DT_REPORT        = C.LIBUSB_DT_REPORT
)

// HID Class-Specific Requests values.
// See section 7.2 of the HID specifications
// https://www.usb.org/document-library/device-class-definition-hid-111
const (
	HID_GET_REPORT         = 0x01
	HID_SET_REPORT         = 0x09
	HID_REPORT_TYPE_INPUT  = 0x01
	HID_REPORT_TYPE_OUTPUT = 0x02
)

// HID magic numbers passed to libusb_control_transfer
const (
	HID_MAGIC_0x0   = 0x0
	HID_MAGIC_0x300 = 0x300
)

// MP707 mysterious protocol commands
// (maybe some of them are 1-Wire actually, but we don't know for sure)
const (
	CMD_MAGIC    = 0x18
	CMD_RESET    = 0x48
	CMD_RW_BIT   = 0x81
	CMD_RW_2BIT  = 0x82
	CMD_RW_4BYTE = 0x84
	CMD_RW_BYTE  = 0x88
)

func sleep(d time.Duration) {
	time.Sleep(d)
}

// InitLib initializes libusb library
func InitLib() error {
	if v := C.libusb_init(nil); v != 0 {
		return errors.New("failed to initialize libusb")
	}

	return nil
}

// DesposeLib destroys libusb library context
func DesposeLib() {
	C.libusb_exit(nil)
}

// extractDesc returns USB device properties reported by libusb
func extractDesc(dev *C.libusb_device) (id uint8, vendor int, product int, err *LibUsbError) {
	desc := &C.struct_libusb_device_descriptor{}

	if r := C.libusb_get_device_descriptor(dev, desc); r != C.int(0) {
		return 0, 0, 0, MakeLibUsbError(C.int(r))
	}

	return uint8(desc.iSerialNumber), int(desc.idVendor), int(desc.idProduct), nil
}

// Lookup returns a list of available USB devices to work with
func Lookup() (devices []Device, ue *LibUsbError) {
	var li **C.struct_libusb_device

	n := C.libusb_get_device_list(nil, &li)
	if n <= 0 {
		return nil, MakeLibUsbError(C.int(n))
	}

	defer C.libusb_free_device_list(li, 1)

	for i := 0; i < int(n); i++ {
		if id, vendor, product, err := extractDesc(C.get_dev(li, C.int(i))); err != nil {
			ue = err
			break
		} else if vendor == VendorId && product == ProductId {
			var handle *C.struct_libusb_device_handle
			if err := C.libusb_open(C.get_dev(li, C.int(i)), &handle); err != 0 {
				ue = MakeLibUsbError(err)
				continue
			}

			devices = append(devices, Device{id: id, handle: handle})
		}
	}

	return
}

// Temperature represents temperature value in degree Celsius
type Temperature float32

func (t Temperature) String() string {
	return fmt.Sprintf("%.2f °C", float32(t))
}

// Device represents MP707 USB device
type Device struct {
	id           uint8
	handle       *C.struct_libusb_device_handle
	outputBuffer [CmdSize]byte
	inputBuffer  [CmdSize]byte
}

// GetId returns device ID
func (d *Device) GetId() uint8 {
	return d.id
}

// clearBuffers zeroes input and output device buffers
func (d *Device) clearBuffers() {
	for i := 0; i < CmdSize; i++ {
		d.inputBuffer[i] = 0
		d.outputBuffer[i] = 0
	}
}

func (d *Device) compareBuffers() error {
	return d.compareBuffersWithDepth(CmdSize)
}

func (d *Device) compareBuffersWithDepth(depth int) error {
	for i := 0; i < depth; i++ {
		if d.inputBuffer[i] != d.outputBuffer[i] {
			return InputOutputError{}
		}
	}

	return nil
}

func (d *Device) usbWriteCtl() *LibUsbError {
	if written := C.libusb_control_transfer(d.handle,
		LIBUSB_ENDPOINT_OUT|LIBUSB_RECIPIENT_DEVICE|LIBUSB_DT_REPORT|HID_REPORT_TYPE_OUTPUT,
		HID_SET_REPORT,
		HID_MAGIC_0x300,
		HID_MAGIC_0x0,
		(*C.uchar)(&d.outputBuffer[0]), C.uint16_t(len(d.outputBuffer)),
		Timeout); int(written) != len(d.outputBuffer) {
		return MakeLibUsbError(written)
	}

	return nil
}

func (d *Device) usbReadCtl() *LibUsbError {
	if read := C.libusb_control_transfer(d.handle,
		LIBUSB_ENDPOINT_IN|LIBUSB_RECIPIENT_DEVICE|LIBUSB_DT_REPORT|HID_REPORT_TYPE_INPUT,
		HID_GET_REPORT,
		HID_MAGIC_0x300,
		HID_MAGIC_0x0,
		(*C.uchar)(&d.inputBuffer[0]), C.uint16_t(len(d.inputBuffer)),
		Timeout); int(read) != len(d.inputBuffer) {
		return MakeLibUsbError(read)
	}

	return nil
}

// usbWriteReadCtl initiates write operation followed by subsequent read operation after an appropriate time delay
func (d *Device) usbWriteReadCtl(delay time.Duration) error {
	if err := d.usbWriteCtl(); err != nil {
		return err
	}

	sleep(delay)

	if err := d.usbReadCtl(); err != nil {
		return err
	}

	return nil
}

// reset sends 'reset' command to device (and probably to 1-Wire slaves)
// The actual command use is unknown, but it seems that one should send 'reset'
// before doing any write/read operations to ensure device readiness
func (d *Device) reset() error {
	d.clearBuffers()

	d.outputBuffer[0] = CMD_MAGIC
	d.outputBuffer[1] = CMD_RESET

	if err := d.usbWriteReadCtl(DELAY_SINGLE); err != nil {
		return err
	}

	if err := d.compareBuffers(); err != nil {
		return err
	}

	return nil
}

// writeBit writes a single bit
func (d *Device) writeBit(bit byte) error {
	d.clearBuffers()

	d.outputBuffer[0] = CMD_MAGIC
	d.outputBuffer[1] = CMD_RW_BIT
	d.outputBuffer[2] = bit & 0x01

	if err := d.usbWriteReadCtl(DELAY_SINGLE); err != nil {
		return err
	}

	if err := d.compareBuffers(); err != nil {
		return err
	}

	return nil
}

// writeByte writes 1 byte
func (d *Device) writeByte(b byte) error {
	d.clearBuffers()

	d.outputBuffer[0] = CMD_MAGIC
	d.outputBuffer[1] = CMD_RW_BYTE
	d.outputBuffer[2] = b

	if err := d.usbWriteReadCtl(DELAY_SINGLE); err != nil {
		return err
	}

	if err := d.compareBuffers(); err != nil {
		return err
	}

	return nil
}

// write4Byte writes 4 bytes
func (d *Device) write4Byte(b uint32) error {
	b0 := b & 0xff
	b1 := (b >> 8) & 0xff
	b2 := (b >> 16) & 0xff
	b3 := (b >> 24) & 0xff

	d.clearBuffers()

	d.outputBuffer[0] = CMD_MAGIC
	d.outputBuffer[1] = CMD_RW_4BYTE
	d.outputBuffer[2] = byte(b0)
	d.outputBuffer[3] = byte(b1)
	d.outputBuffer[4] = byte(b2)
	d.outputBuffer[5] = byte(b3)

	if err := d.usbWriteReadCtl(DELAY_DOUBLE); err != nil {
		return err
	}

	if err := d.compareBuffers(); err != nil {
		return err
	}

	return nil
}

// read2Bit reads 2 bits
func (d *Device) read2Bit() (bit byte, err error) {
	d.clearBuffers()

	d.outputBuffer[0] = CMD_MAGIC
	d.outputBuffer[1] = CMD_RW_2BIT
	d.outputBuffer[2] = 0x01
	d.outputBuffer[3] = 0x01

	if err := d.usbWriteReadCtl(DELAY_SINGLE); err != nil {
		return 0, err
	}

	bit = d.inputBuffer[2] + d.inputBuffer[3]<<1

	if err := d.compareBuffersWithDepth(2); err != nil {
		return bit, err
	}

	return
}

// ReadByte reads 1 byte
func (d *Device) readByte() (bit byte, err error) {
	d.clearBuffers()

	d.outputBuffer[0] = CMD_MAGIC
	d.outputBuffer[1] = CMD_RW_BYTE
	d.outputBuffer[2] = 0xff

	if err := d.usbWriteReadCtl(DELAY_SINGLE); err != nil {
		return 0, err
	}

	bit = d.inputBuffer[2]

	if err := d.compareBuffersWithDepth(2); err != nil {
		return bit, err
	}

	return
}

// read4Byte reads 4 bytes
func (d *Device) read4Byte() (b uint32, err error) {
	d.clearBuffers()

	d.outputBuffer[0] = CMD_MAGIC
	d.outputBuffer[1] = CMD_RW_4BYTE
	d.outputBuffer[2] = 0xff
	d.outputBuffer[3] = 0xff
	d.outputBuffer[4] = 0xff
	d.outputBuffer[5] = 0xff

	if err := d.usbWriteReadCtl(DELAY_DOUBLE); err != nil {
		return 0, err
	}

	b = uint32(d.inputBuffer[2]) +
		uint32(d.inputBuffer[3])<<8 +
		uint32(d.inputBuffer[4])<<16 +
		uint32(d.inputBuffer[5])<<24

	if err := d.compareBuffersWithDepth(2); err != nil {
		return b, err
	}

	return
}

// searchRom performs a 'Search ROM' procedure and returns ROMs of available devices
func (d *Device) searchRom(romNext onewire.Rom, portLast uint32) (roms []onewire.Rom, err error) {
	ok := false
	cl := make([]bool, RomPorts)
	rl := make([]onewire.Rom, RomPorts)

	var bit uint8
	var crc onewire.Crc
	var rom onewire.Rom

	for n := 0; n < Retries && !ok; n++ {
		if err := d.reset(); err != nil {
			continue
		}

		if err := d.writeByte(onewire.SearchRom); err != nil {
			continue
		}

		ok = true

		if ok {
			for i := 0; i < RomPorts; i++ {
				if ok {
					if bit, err = d.read2Bit(); err == nil {
						switch bit & 0x03 {
						case 0: // collision?
							if portLast < uint32(i) {
								cl[i] = true
								rl[i] = rom
								bit = 0
							} else { // portLast >= i
								bit = uint8(romNext >> i & 1)
							}

							if err := d.writeBit(bit); err != nil {
								ok = false
								break
							}

							if bit == 1 {
								rom += 1 << i
							}

						case 1:
							if err := d.writeBit(1); err != nil {
								ok = false
								break
							} else {
								rom += 1 << i
							}
						case 2:
							if err := d.writeBit(0); err != nil {
								ok = false
								break
							}
						case 3:
							ok = false
							break
						}
					} else { // if read2Bit
						ok = false
						break
					}
				} // if ok
			} // for i..RomPorts
		} // if ok

		if rom == 0 {
			ok = false
		}

		if ok {
			crc = 0
			for j := 0; j < 8; j++ {
				crc = onewire.Crc8(byte(crc), byte(rom>>(j<<3))&0xff)

				if err := onewire.CheckCrc8(crc); err != nil {
					ok = false
				} else {
					ok = true
				}
			}
		}
	} // for n..Retries

	if !ok {
		return
	} else {
		roms = append(roms, rom)
	}

	for i := 0; i < RomPorts; i++ {
		if cl[i] {
			if r, e := d.searchRom(rl[i]|1<<i, uint32(i)); e == nil {
				roms = append(roms, r...)
			}
		}
	}

	if !ok {
		err = errors.New("failed to enumerate sensors")
	}

	return
}

// matchRom issues 'Match ROM' command to address subsequent command to a specific slave device
func (d *Device) matchRom(rom onewire.Rom) (err error) {
	if err := d.reset(); err != nil {
		return err
	}

	if err := d.writeByte(onewire.MatchRom); err != nil {
		return err
	}

	if err := d.write4Byte(uint32(rom & 0xFFFFFFFF)); err != nil {
		return err
	}

	if err := d.write4Byte(uint32((rom >> 32) & 0xFFFFFFFF)); err != nil {
		return err
	}

	return
}

func (d *Device) skipRomConvert() (err error) {
	if err := d.reset(); err != nil {
		return err
	}

	if err := d.writeByte(onewire.SkipRom); err != nil {
		return err
	}

	if err := d.writeByte(onewire.ConvertT); err != nil {
		return err
	}

	return
}

// GetSensors returns a slice of ROMs of connected 1-Wire temperature sensors
func (d *Device) GetSensors() (roms []onewire.Rom, err error) {
	roms, err = d.searchRom(0, 0)
	return
}

// GetTemperature reads temperature sensor value by its ROM address
func (d *Device) GetTemperature(rom onewire.Rom) (temperature Temperature, err error) {
	var l1, l2 uint32
	var l3 byte

	for n := 0; n < Retries; n++ {
		if err := d.skipRomConvert(); err != nil {
			continue
		}
		if err := d.matchRom(rom); err != nil {
			continue
		}

		if err := d.writeByte(onewire.ReadScratchpad); err != nil {
			continue
		}

		if l1, err = d.read4Byte(); err != nil {
			continue
		}

		if l2, err = d.read4Byte(); err != nil {
			continue
		}

		if l3, err = d.readByte(); err != nil {
			continue
		}

		var crc onewire.Crc

		for i := 0; i < 4; i++ {
			crc = onewire.Crc8(byte(crc), byte((l1>>(i*8))&0xff))
		}
		for i := 0; i < 4; i++ {
			crc = onewire.Crc8(byte(crc), byte((l2>>(i*8))&0xff))
		}
		crc = onewire.Crc8(byte(crc), l3)

		if err := onewire.CheckCrc8(crc); err != nil {
			continue
		}

		k := float32(int16(l1 & 0xffff))

		switch onewire.GetFamilyByRom(rom) {
		case onewire.FamilyDS18B20:
			fallthrough
		case onewire.FamilyDS1822:
			temperature = Temperature(k * .0625)
		case onewire.FamilyDS18S20:
			temperature = Temperature(k * .5)
		default:
			temperature = 1000
		}

		return

	}

	return 0, InputOutputError{}
}
