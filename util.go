package dht

import (
	"crypto/rand"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// randomString generates a size-length string randomly.
func randomString(size int) string {
	buff := make([]byte, size)
	rand.Read(buff)
	return string(buff)
}

// bytes2int returns the int value it represents.
func bytes2int(data []byte) uint64 {
	n, val := len(data), uint64(0)
	if n > 8 {
		panic("data too long")
	}

	for i, b := range data {
		val += uint64(b) << uint64((n-i-1)*8)
	}
	return val
}

// int2bytes returns the byte array it represents.
func int2bytes(val uint64) []byte {
	data, j := make([]byte, 8), -1
	for i := 0; i < 8; i++ {
		shift := uint64((7 - i) * 8)
		data[i] = byte((val & (0xff << shift)) >> shift)

		if j == -1 && data[i] != 0 {
			j = i
		}
	}

	if j != -1 {
		return data[j:]
	}
	return data[:1]
}

// decodeCompactIPPortInfo decodes compactIP-address/port info in BitTorrent
// DHT Protocol. It returns the ip and port number.
func decodeCompactIPPortInfo(info string) (ip net.IP, port int, err error) {
	if len(info) != 6 && len(info) != 18 {
		err = errors.New("compact info should be 6-length (IPv4) or 18-length (IPv6) long")
		return
	}

	if len(info) == 6 {
		// IPv4
		ip = net.IPv4(info[0], info[1], info[2], info[3])
		port = int((uint16(info[4]) << 8) | uint16(info[5]))
	} else {
		// IPv6
		ip = net.IP(info[:16])
		port = int((uint16(info[16]) << 8) | uint16(info[17]))
	}
	return
}

// encodeCompactIPPortInfo encodes an ip and a port number to
// compactIP-address/port info.
func encodeCompactIPPortInfo(ip net.IP, port int) (info string, err error) {
	if port > 65535 || port < 0 {
		err = errors.New(
			"port should be no greater than 65535 and no less than 0")
		return
	}

	// Convert to 4-byte or 16-byte IP representation
	ip4 := ip.To4()
	if ip4 != nil {
		// IPv4
		p := int2bytes(uint64(port))
		if len(p) < 2 {
			p = append(p, p[0])
			p[0] = 0
		}
		info = string(append(ip4, p...))
	} else {
		// IPv6
		ip16 := ip.To16()
		if ip16 == nil {
			err = errors.New("invalid IP address")
			return
		}
		p := make([]byte, 2)
		p[0] = byte(port >> 8)
		p[1] = byte(port)
		info = string(append(ip16, p...))
	}
	return
}

// getLocalIPs returns local ips.
func getLocalIPs() (ips []string) {
	ips = make([]string, 0, 6)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		ips = append(ips, ip.String())
	}
	return
}

// getRemoteIP returns the wlan ip.
func getRemoteIP() (ip string, err error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest("GET", "http://ifconfig.me", nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "curl")
	res, err := client.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	ip = string(data)

	return
}

// genAddress returns a ip:port address.
func genAddress(ip string, port int) string {
	return strings.Join([]string{ip, strconv.Itoa(port)}, ":")
}
