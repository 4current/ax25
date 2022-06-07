package main

import (
	"errors"
	"fmt"
	"io"
	"regexp"
)

const (
	bit0 = byte(1) << iota
	_
	_
	_
	_
	bit5
	bit6
	bit7
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

//func frame(buffer *bytes.Buffer, payload string, withStartFlag bool) {
//	if withStartFlag {
//		buffer.WriteByte(0x7e)
//	}
//	buffer.WriteString(payload)
//	buffer.WriteByte(0x7e)
//
//}

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
		addr[A14] &= bit7 ^ 0xFF
	} else {
		addr[A7] &= bit7 ^ 0xFF
		addr[A14] |= bit7
	}

	// if more address octets follow unset bit0 else set bit0
	if moreAddr {
		addr[A14] &= bit0 ^ 0xFF
	} else {
		addr[A14] |= bit0
	}

	return addr
}

func main() {
}
