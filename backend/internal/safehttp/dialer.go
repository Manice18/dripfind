package safehttp

import (
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

func Control(network, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("unresolved address %q", address)
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return fmt.Errorf("blocked address %s", ip)
	}
	return nil
}

func Transport() *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{Timeout: 10 * time.Second, Control: Control}).DialContext,
	}
}
