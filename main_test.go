package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_addressEncode(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name     string
		args     args
		wantCall string
		wantSsid int
		wantErr  error
	}{
		{
			name: "Just a call sign",
			args: args{
				address: "WWV",
			},
			wantCall: "WWV",
			wantSsid: 0,
			wantErr:  nil,
		},
		{
			name: "call sign ending in a number",
			args: args{
				address: "WW4V2",
			},
			wantCall: "WW4V2",
			wantSsid: 0,
			wantErr:  nil,
		},
		{
			name: "call sign too large",
			args: args{
				address: "WW4V2XXX",
			},
			wantCall: "",
			wantSsid: 0,
			wantErr:  invalidAddressError,
		},

		{
			name: "A call sign with ssid",
			args: args{
				address: "WWV-1",
			},
			wantCall: "WWV",
			wantSsid: 1,
			wantErr:  nil,
		},
		{
			name: "A call sign with max ssid",
			args: args{
				address: "WWV-15",
			},
			wantCall: "WWV",
			wantSsid: 15,
			wantErr:  nil,
		},
		{
			name: "A call sign with invalid SSID",
			args: args{
				address: "WWV-16",
			},
			wantCall: "",
			wantSsid: 0,
			wantErr:  invalidAddressError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCall, gotSsid, gotErr := addressEncode(tt.args.address)
			assert.Equal(t, tt.wantCall, gotCall)
			assert.Equal(t, tt.wantSsid, gotSsid)
			assert.Equal(t, tt.wantErr, gotErr)
		})
	}
}

func Test_encAddr(t *testing.T) {
	type args struct {
		src       string
		dst       string
		isCommand bool
		moreAddr  bool
	}
	tests := []struct {
		name     string
		args     args
		wantAddr [14]byte
	}{
		{
			name: "spec2.2 example - command",
			args: args{
				src:       "N7LEM",
				dst:       "NJ7P",
				isCommand: true,
				moreAddr:  false,
			},
			wantAddr: [14]byte{0x9c, 0x94, 0x6e, 0xa0, 0x40, 0x40, 0xe0, 0x9c, 0x6e, 0x98, 0x8a, 0x9A, 0x40, 0x61},
		},
		{
			name: "spec2.2 example - response",
			args: args{
				src:       "N7LEM",
				dst:       "NJ7P",
				isCommand: false,
				moreAddr:  false,
			},
			wantAddr: [14]byte{0x9c, 0x94, 0x6e, 0xa0, 0x40, 0x40, 0x60, 0x9c, 0x6e, 0x98, 0x8a, 0x9A, 0x40, 0xe1},
		},
		{
			name: "with ssid - command",
			args: args{
				src:       "N7LEM-4",
				dst:       "NJ7P",
				isCommand: true,
				moreAddr:  false,
			},
			wantAddr: [14]byte{0x9c, 0x94, 0x6e, 0xa0, 0x40, 0x40, 0xe0, 0x9c, 0x6e, 0x98, 0x8a, 0x9A, 0x40, 0x69},
		},
		{
			name: "with ssid - response",
			args: args{
				src:       "N7LEM",
				dst:       "NJ7P-15",
				isCommand: false,
				moreAddr:  false,
			},
			wantAddr: [14]byte{0x9c, 0x94, 0x6e, 0xa0, 0x40, 0x40, 0x7e, 0x9c, 0x6e, 0x98, 0x8a, 0x9A, 0x40, 0xe1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAddr := encAddr(tt.args.src, tt.args.dst, tt.args.isCommand, tt.args.moreAddr)
			assert.EqualValues(t, gotAddr, tt.wantAddr)
		})
	}
}
