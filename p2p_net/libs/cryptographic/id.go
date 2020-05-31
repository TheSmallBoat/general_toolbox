package cryptographic

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"
	"strconv"
	"strings"

	"awesomeProject/beacon/p2p_network/libs/common"
)

// ID represents a peer ID. It comprises of a cryptographic public key, and a public, reachable network address
// specified by a IPv4/IPv6 host and 16-bit port number. The size of an ID in terms of its byte representation
// is static, with its contents being deterministic.
type ID struct {
	// The Ed25519 public key of the bearer of this ID.
	PubKey PublicKey `json:"public_key"`

	// Public host of the bearer of this ID.
	Host net.IP `json:"address"`

	// Public port of the bearer of this ID.
	Port uint16

	// 'host:port'
	Address string
}

// NewID instantiates a new, immutable cryptographic user ID.
func NewID(pubKey PublicKey, host net.IP, port uint16) ID {
	addr := net.JoinHostPort(common.NormalizeIP(host), strconv.FormatUint(uint64(port), 10))
	return ID{PubKey: pubKey, Host: host, Port: port, Address: addr}
}

// Size returns the number of bytes this ID comprises of.
func (i ID) Size() int {
	return len(i.PubKey) + net.IPv6len + 2
}

// String returns a JSON representation of this ID.
func (i ID) String() string {
	var builder strings.Builder
	builder.WriteString(`{"public_key": "`)
	builder.WriteString(hex.EncodeToString(i.PubKey[:]))
	builder.WriteString(`", "address": "`)
	builder.WriteString(i.Address)
	builder.WriteString(`"}`)
	return builder.String()
}

// Marshal serializes this ID into its byte representation.
func (i ID) Marshal() []byte {
	buf := make([]byte, i.Size())

	copy(buf[:len(i.PubKey)], i.PubKey[:])
	copy(buf[len(i.PubKey):len(i.PubKey)+net.IPv6len], i.Host)
	binary.BigEndian.PutUint16(buf[len(i.PubKey)+net.IPv6len:len(i.PubKey)+net.IPv6len+2], i.Port)

	return buf
}

// UnmarshalID deserializes buf, representing a slice of bytes, ID instance. It throws io.ErrUnexpectedEOF if the
// contents of buf is malformed.
func UnmarshalID(buf []byte) (ID, error) {
	if len(buf) < SizePublicKey {
		return ID{}, io.ErrUnexpectedEOF
	}

	var pubKey PublicKey

	copy(pubKey[:], buf[:SizePublicKey])
	buf = buf[SizePublicKey:]

	if len(buf) < net.IPv6len {
		return ID{}, io.ErrUnexpectedEOF
	}

	host := make([]byte, net.IPv6len)
	copy(host, buf[:net.IPv6len])

	buf = buf[net.IPv6len:]

	if len(buf) < 2 {
		return ID{}, io.ErrUnexpectedEOF
	}

	port := binary.BigEndian.Uint16(buf[:2])

	return NewID(pubKey, host, port), nil
}
