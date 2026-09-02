package devices

import (
	"fmt"
	"net"
	"strconv"
)

func splitHostPort(address string) (host string, port int, err error) {
	host, portValue, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, fmt.Errorf("address %q must be in host:port form", address)
	}
	port, err = strconv.Atoi(portValue)
	if err != nil {
		return "", 0, fmt.Errorf("address %q has a non-numeric port", address)
	}
	if port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("address %q port must be in range 1..65535", address)
	}
	return host, port, nil
}
