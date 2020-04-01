package tests

import (
	"net"
	"time"
)

func PingTcpAddress(address string) error {
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}