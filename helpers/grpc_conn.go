package helpers

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func parseGRpcAddress(addr string) (retSchema, retAddress string, retErr error) {
	// Handling default VM config for previous version
	if !strings.Contains(addr, "://") {
		addr = "tcp://" + addr
	}

	u, err := url.Parse(addr)
	if err != nil {
		retErr = fmt.Errorf("url parse failed: %w", err)
		return
	}
	retSchema = u.Scheme

	// u.Path / u.Host depends on u.Scheme, so we combine them
	retAddress = u.Host + u.Path

	return
}

// Get net.Listener for UNIX/TCP address string.
func GetGRpcNetListener(addr string) (net.Listener, error) {
	schema, address, err := parseGRpcAddress(addr)
	if err != nil {
		return nil, err
	}

	// Remove socket file if exists
	if schema == "unix" {
		os.Remove(address)
	}

	listener, err := net.Listen(schema, address)
	if err != nil {
		return nil, fmt.Errorf("net.Listen failed: %w", err)
	}

	return listener, nil
}

// Get gRPC client connection for UNIX/TCP address string.
// Keep alive option is not added if {keepAlivePeriod} == 0.
func GetGRpcClientConnection(addr string, keepAlivePeriod time.Duration) (*grpc.ClientConn, error) {
	schema, address, err := parseGRpcAddress(addr)
	if err != nil {
		return nil, err
	}

	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			timeout := time.Duration(0)
			if deadline, ok := ctx.Deadline(); ok {
				timeout = time.Until(deadline)
			}

			return net.DialTimeout(schema, address, timeout)
		}),
	}

	if keepAlivePeriod > 0 {
		kpParams := keepalive.ClientParameters{
			Time:                keepAlivePeriod, // send pings every 1 second if there is no activity
			Timeout:             keepAlivePeriod, // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		}

		dialOptions = append(dialOptions, grpc.WithKeepaliveParams(kpParams))
	}

	// Bypass Rust h2 library UDS limitations: uri validation failure causing PROTOCOL_ERROR gRPC error
	dialAddress :=  address
	if schema == "unix" {
		dialAddress = "127.0.0.1" // faking filePath with valid URL
	}

	return grpc.Dial(dialAddress, dialOptions...)
}
