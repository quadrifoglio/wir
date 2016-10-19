package utils

import (
	"net"
)

// IncrementIP increments the specified
// IP address
func IncrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++

		if ip[j] > 0 {
			break
		}
	}
}
