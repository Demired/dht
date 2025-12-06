package dht

import (
	"testing"
)

func TestInt2Bytes(t *testing.T) {
	cases := []struct {
		in  uint64
		out []byte
	}{
		{0, []byte{0}},
		{1, []byte{1}},
		{256, []byte{1, 0}},
		{22129, []byte{86, 113}},
	}

	for _, c := range cases {
		r := int2bytes(c.in)
		if len(r) != len(c.out) {
			t.Fail()
		}

		for i, v := range r {
			if v != c.out[i] {
				t.Fail()
			}
		}
	}
}

func TestBytes2Int(t *testing.T) {
	cases := []struct {
		in  []byte
		out uint64
	}{
		{[]byte{0}, 0},
		{[]byte{1}, 1},
		{[]byte{1, 0}, 256},
		{[]byte{86, 113}, 22129},
	}

	for _, c := range cases {
		if bytes2int(c.in) != c.out {
			t.Fail()
		}
	}
}

func TestDecodeCompactIPPortInfo(t *testing.T) {
	cases := []struct {
		in  string
		out struct {
			ip   string
			port int
		}
	}{
		{"123456", struct {
			ip   string
			port int
		}{"49.50.51.52", 13622}},
		{"abcdef", struct {
			ip   string
			port int
		}{"97.98.99.100", 25958}},
	}

	for _, item := range cases {
		ip, port, err := decodeCompactIPPortInfo(item.in)
		if err != nil || ip.String() != item.out.ip || port != item.out.port {
			t.Fail()
		}
	}

	// Test IPv6 - using raw bytes for IPv6 address 2001:db8::1 with port 6881
	ipv6Data := string([]byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x1a, 0xe1})
	ip, port, err := decodeCompactIPPortInfo(ipv6Data)
	if err != nil {
		t.Errorf("IPv6 decode failed: %v", err)
	}
	if ip.String() != "2001:db8::1" || port != 6881 {
		t.Errorf("IPv6 decode failed: got %s:%d, expected %s:%d", ip.String(), port, "2001:db8::1", 6881)
	}
}

func TestEncodeCompactIPPortInfo(t *testing.T) {
	cases := []struct {
		in struct {
			ip   []byte
			port int
		}
		out string
	}{
		{struct {
			ip   []byte
			port int
		}{[]byte{49, 50, 51, 52}, 13622}, "123456"},
		{struct {
			ip   []byte
			port int
		}{[]byte{97, 98, 99, 100}, 25958}, "abcdef"},
	}

	for _, item := range cases {
		info, err := encodeCompactIPPortInfo(item.in.ip, item.in.port)
		if err != nil || info != item.out {
			t.Fail()
		}
	}

	// Test IPv6 - manually create the expected bytes for IPv6 address 2001:db8::1 with port 6881
	ip := []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	port := 6881

	info, err := encodeCompactIPPortInfo(ip, port)
	if err != nil {
		t.Errorf("IPv6 encode failed: %v", err)
	}

	expected := []byte{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x1a, 0xe1}
	if len(info) != 18 {
		t.Errorf("IPv6 encoded info wrong length: got %d, expected 18", len(info))
	}

	for i, b := range []byte(info) {
		if b != expected[i] {
			t.Errorf("IPv6 encoded info mismatch at byte %d: got %x, expected %x", i, b, expected[i])
		}
	}
}
