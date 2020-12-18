package onewire

// CrcMismatchError represents CRC verification error
type CrcMismatchError struct{}

func (cme CrcMismatchError) Error() string {
	return "crc mismatch error"
}
