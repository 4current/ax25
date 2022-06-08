package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sigurn/crc16"
	"io"
	"regexp"
)

const (
	bit0, mask0 = byte(1) << iota, byte(1)<<iota ^ 0xFF // bit0 == 0x01, mask0 == 0xFE  (iota == 0)
	_, _                                                //                              (iota == 1, unused)
	_, _                                                //                              (iota == 2, unused)
	_, _                                                //                              (iota == 3, unused)
	bit4, mask4                                         // bit4 == 0x10, mask4 == 0xEF  (iota == 4)
	bit5, _                                             // bit5 == 0x20, mask5 == 0xDF  (iota == 5)
	bit6, _                                             // bit6 == 0x40, mask6 == 0xBF  (iota == 6)
	bit7, mask7                                         // bit7 == 0x80, mask7 == 0x7F  (iota == 7)
)

const (
	A1 = iota
	_
	_
	_
	_
	A6
	A7
	A8
	_
	_
	_
	_
	A13
	A14
)

const (
	FlagField byte = 0x7e
	PidField  byte = 0xF0
)

var (
	sndSeqNumber  = 1
	rcvSeqNumber  = 1
	pollFinalFlag = true
	fcsTable      = crc16.MakeTable(crc16.CRC16_X_25)
)

func getFcs(frame []byte) []byte {
	fcsBytes := make([]byte, 2)
	fcs := crc16.Checksum(frame, fcsTable)
	binary.BigEndian.PutUint16(fcsBytes, fcs)
	return fcsBytes
}

func controlField(rcvSeqNum int, sndSeqNum int, pollFinalFlag bool) (ctrlField byte) {

	if rcvSeqNum >= 0 && rcvSeqNum < 8 {
		ctrlField = byte(rcvSeqNum << 5)
	}
	if pollFinalFlag {
		ctrlField |= bit4
	} else {
		ctrlField &= mask4
	}
	if sndSeqNum >= 0 && sndSeqNum < 8 {
		ctrlField = byte(sndSeqNum) << 1
	}
	return
}

func buildFrame(buffer *bytes.Buffer, addressField [14]byte, message string, withStartFlag bool) {
	if withStartFlag {
		buffer.WriteByte(FlagField)
	}
	payload := buffer.Next(1)
	buffer.Write(addressField[:])
	buffer.WriteByte(controlField(rcvSeqNumber, sndSeqNumber, pollFinalFlag))
	buffer.WriteByte(PidField)
	buffer.WriteString(message)
	payload = buffer.Next(len(message) + 16)
	buffer.Write(getFcs(payload))
	buffer.WriteByte(FlagField)

}

var invalidAddressError = errors.New("the ax.25 address is invalid")

func addressEncode(address string) (call string, ssid int, err error) {
	pat, _ := regexp.Compile(`(^[A-Z\d]{2,6})(-(\d|1[0-5]))?$`)
	pat.Longest()
	subs := pat.FindStringSubmatch(address)
	if subs == nil {
		err = invalidAddressError
		return
	}
	call = subs[1]
	_, err = fmt.Sscan(subs[3], &ssid)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return
}

func encAddr(src string, dst string, isCommand bool, moreAddr bool) (addr [14]byte) {

	// copy in dst call sign and SSID
	addrDst, ssidDst, _ := addressEncode(dst)
	i := copy(addr[A1:A6], addrDst)
	for ; i < A7; i++ {
		addr[i] = ' '
	}
	addr[A7] = byte(ssidDst)

	// copy in src call sign and SSID
	addrSrc, ssidSrc, _ := addressEncode(src)
	j := copy(addr[A8:A13], addrSrc)
	for j += A8; j < A14; j++ {
		addr[j] = ' '
	}
	addr[A14] = byte(ssidSrc)

	// left shift for extension bits
	for k := A1; k <= A14; k++ {
		addr[k] <<= 1
	}

	// set reserved bits
	addr[A7] |= bit5 + bit6
	addr[A14] |= bit5 + bit6

	// set command / response flags
	if isCommand {
		addr[A7] |= bit7
		addr[A14] &= mask7
	} else {
		addr[A7] &= mask7
		addr[A14] |= bit7
	}

	// if more address octets follow unset bit0 else set bit0
	if moreAddr {
		addr[A14] &= mask0
	} else {
		addr[A14] |= bit0
	}

	return addr
}

func main() {
	var aprsMessage bytes.Buffer
	addresses := encAddr("AE4OK-1", "APRX29", true, false)
	buildFrame(&aprsMessage, addresses, "hello", true)
	fmt.Print(aprsMessage)
}
