package ceph

import (
	"reflect"
	"testing"
)

func TestParseAddresses(t *testing.T) {
	t.Parallel()

	cephExample := "[v2:10.97.145.7:3300,v1:10.97.145.7:6789],[v2:10.97.167.34:3300,v1:10.97.167.34:6789],[v2:10.97.166.34:3300,v1:10.97.166.34:6789]"
	wantCeph := [][]address{
		{
			{addrType: addrTypeV2, host: "10.97.145.7", port: 3300},
			{addrType: addrTypeV1, host: "10.97.145.7", port: 6789},
		},
		{
			{addrType: addrTypeV2, host: "10.97.167.34", port: 3300},
			{addrType: addrTypeV1, host: "10.97.167.34", port: 6789},
		},
		{
			{addrType: addrTypeV2, host: "10.97.166.34", port: 3300},
			{addrType: addrTypeV1, host: "10.97.166.34", port: 6789},
		},
	}

	tests := []struct {
		name    string
		input   string
		want    [][]address
		wantErr bool
	}{
		{name: "ceph bracketed three monitors", input: cephExample, want: wantCeph},
		{
			name:  "unbracketed single monitor",
			input: "v2:10.97.145.7:3300,v1:10.97.145.7:6789",
			want: [][]address{
				{
					{addrType: addrTypeV2, host: "10.97.145.7", port: 3300},
					{addrType: addrTypeV1, host: "10.97.145.7", port: 6789},
				},
			},
		},
		{
			name:  "unbracketed single monitor with nonce",
			input: "v2:10.0.0.10:3300/0,v1:10.0.0.10:6789/0",
			want: [][]address{
				{
					{addrType: addrTypeV2, host: "10.0.0.10", port: 3300, nonce: 0},
					{addrType: addrTypeV1, host: "10.0.0.10", port: 6789, nonce: 0},
				},
			},
		},
		{
			name:  "flat two monitors same-host merge",
			input: "v2:10.0.0.1:3300,v1:10.0.0.1:6789,v2:10.0.0.2:3300,v1:10.0.0.2:6789",
			want: [][]address{
				{
					{addrType: addrTypeV2, host: "10.0.0.1", port: 3300},
					{addrType: addrTypeV1, host: "10.0.0.1", port: 6789},
				},
				{
					{addrType: addrTypeV2, host: "10.0.0.2", port: 3300},
					{addrType: addrTypeV1, host: "10.0.0.2", port: 6789},
				},
			},
		},
		{
			name:  "single bracket group",
			input: "[v2:192.0.2.1:3300,v1:192.0.2.1:6789]",
			want: [][]address{
				{
					{addrType: addrTypeV2, host: "192.0.2.1", port: 3300},
					{addrType: addrTypeV1, host: "192.0.2.1", port: 6789},
				},
			},
		},
		{
			name:  "bracketed ipv6 single addr no comma no strip",
			input: "[2001:db8::1]:3300",
			want: [][]address{
				{{addrType: addrTypeAny, host: "2001:db8::1", port: 3300}},
			},
		},
		{name: "empty", input: "", want: nil},
		{name: "whitespace only", input: "  \t\n", want: nil},
		{name: "invalid token", input: "v2:10.0.0.1:badport", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseAddresses(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil, result=%v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestParseMonitorAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    address
		wantErr bool
	}{
		{
			name:  "full format with v2 and nonce",
			input: "v2:10.0.0.10:3300/0",
			want: address{
				addrType: addrTypeV2,
				host:     "10.0.0.10",
				port:     3300,
				nonce:    0,
			},
		},
		{
			name:  "v1 host without port and nonce",
			input: "v1:mon.example.com",
			want: address{
				addrType: addrTypeV1,
				host:     "mon.example.com",
			},
		},
		{
			name:  "host and port without type",
			input: "10.0.0.11:6789",
			want: address{
				addrType: addrTypeAny,
				host:     "10.0.0.11",
				port:     6789,
			},
		},
		{
			name:  "bracketed ipv6 with port and nonce",
			input: "[2001:db8::1]:3300/12",
			want: address{
				addrType: addrTypeAny,
				host:     "2001:db8::1",
				port:     3300,
				nonce:    12,
			},
		},
		{
			name:  "unbracketed ipv6 without port",
			input: "2001:db8::1",
			want: address{
				addrType: addrTypeAny,
				host:     "2001:db8::1",
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid port",
			input:   "10.0.0.10:abc",
			wantErr: true,
		},
		{
			name:    "empty nonce",
			input:   "10.0.0.10:3300/",
			wantErr: true,
		},
		{
			name:    "negative nonce",
			input:   "10.0.0.10:3300/-1",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseAddress(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result=%+v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected parse result: got=%+v want=%+v", got, tc.want)
			}
		})
	}
}
