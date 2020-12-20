package usb

/*
 #cgo pkg-config: libusb-1.0
 #include <libusb-1.0/libusb.h>
*/
import "C"
import "fmt"

// LibUsbErrorCode is an error number reported by libusb
type LibUsbErrorCode int

// LibUsbErrorDesc is an error description reported by libusb
type LibUsbErrorDesc string

// LibUsbError represents libusb error
type LibUsbError struct {
	code LibUsbErrorCode
	desc LibUsbErrorDesc
}

// MakeLibUsbError creates new LibUsbError with filled in code and description struct fields
func MakeLibUsbError(code C.int) *LibUsbError {
	return &LibUsbError{
		LibUsbErrorCode(code),
		LibUsbErrorDesc(C.GoString(C.libusb_error_name(code)))}
}

func (ue LibUsbError) Error() string {
	return fmt.Sprintf("Error(%d): %s", ue.code, ue.desc)
}

// InputOutputError represents general i/o error
type InputOutputError struct{}

func (ioe InputOutputError) Error() string {
	return "input/output error"
}
