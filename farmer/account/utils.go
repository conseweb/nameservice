package account

import (
	"net"
	"strings"
)

func checkPhone(phone string) bool {
	return phone == ""
}

func checkEmail(email string) bool {
	return email == ""
}

func getLocalMAC() string {
	list, err := net.Interfaces()
	if err != nil {
		return ""
	}

SKIP_DEV:
	for _, dev := range list {
		for _, v := range []string{"lo", "br", "etch"} {
			if strings.HasPrefix(dev.Name, v) {
				continue SKIP_DEV
			}
		}

		addrs, err := dev.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if len(strings.Split(addr.String(), ".")) == 4 {
				return dev.HardwareAddr.String()
			}
		}
	}
	return ""
}
