package main

import (
	"errors"
	"fmt"
	"gl-socks5-demo/common"
	"io"
	"net"
	"time"
)

const Name = "socks5"

// https://www.ietf.org/rfc/rfc1928.txt

// Version is socks5 version number.
const Version5 = 0x05

// SOCKS auth type
const (
	AuthNone     = 0x00
	AuthPassword = 0x02
)

// SOCKS request commands as defined in RFC 1928 section 4
const (
	CmdConnect      = 0x01
	CmdBind         = 0x02
	CmdUDPAssociate = 0x03
)

// SOCKS address types as defined in RFC 1928 section 4
const (
	ATypIP4    = 0x1
	ATypDomain = 0x3
	ATypIP6    = 0x4
)

func Handshake(underlay net.Conn) (io.ReadWriter, *TargetAddr, error) {
	// Set handshake timeout 4 seconds
	if err := underlay.SetReadDeadline(time.Now().Add(time.Second * 4)); err != nil {
		return nil, nil, err
	}
	defer underlay.SetReadDeadline(time.Time{})

	// https://www.ietf.org/rfc/rfc1928.txt
	buf := common.GetBuffer(512)
	defer common.PutBuffer(buf)

	// Read hello message
	n, err := underlay.Read(buf)
	if err != nil || n == 0 {
		return nil, nil, fmt.Errorf("failed to read hello: %w", err)
	}
	version := buf[0]
	if version != Version5 {
		return nil, nil, fmt.Errorf("unsupported socks version %v", version)
	}

	// Write hello response
	// TODO: Support Auth
	_, err = underlay.Write([]byte{Version5, AuthNone})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write hello response: %w", err)
	}

	// Read command message
	n, err = underlay.Read(buf)
	if err != nil || n < 7 { // Shortest length is 7
		return nil, nil, fmt.Errorf("failed to read command: %w", err)
	}
	cmd := buf[1]
	if cmd != CmdConnect {
		return nil, nil, fmt.Errorf("unsuppoted command %v", cmd)
	}
	addr := TargetAddr{}
	l := 2
	off := 4
	switch buf[3] {
	case ATypIP4:
		l += net.IPv4len
		addr.IP = make(net.IP, net.IPv4len)
	case ATypIP6:
		l += net.IPv6len
		addr.IP = make(net.IP, net.IPv6len)
	case ATypDomain:
		l += int(buf[4])
		off = 5
	default:
		return nil, nil, fmt.Errorf("unknown address type %v", buf[3])
	}

	if len(buf[off:]) < l {
		return nil, nil, errors.New("short command request")
	}
	if addr.IP != nil {
		copy(addr.IP, buf[off:])
	} else {
		addr.Name = string(buf[off : off+l-2])
	}
	addr.Port = int(buf[off+l-2])<<8 | int(buf[off+l-1])

	// Write command response
	_, err = underlay.Write([]byte{Version5, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write command response: %w", err)
	}

	return underlay, &addr, err
}
