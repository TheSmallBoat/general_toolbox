package cryptographic

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
)

func TestID_String(t *testing.T) {
	t.Parallel()

	f := func(pubKey PublicKey, host net.IP, port uint16) bool {
		if host.IsLoopback() || host.IsUnspecified() { // Make-shift 'NormalizeIP(net.IP)'.
			host = nil
		}

		h := host.String()
		if h == "<nil>" {
			h = ""
		}

		id := NewID(pubKey, host, port)

		if !assert.Equal(t,
			fmt.Sprintf(
				`{"public_key": "%s", "address": "%s"}`,
				pubKey, net.JoinHostPort(h, strconv.FormatUint(uint64(port), 10)),
			),
			id.String(),
		) {
			return false
		}

		return true
	}

	assert.NoError(t, quick.Check(f, nil))
}

func TestUnmarshalID(t *testing.T) {
	t.Parallel()

	_, err := UnmarshalID(nil)
	assert.EqualError(t, err, io.ErrUnexpectedEOF.Error())

	_, err = UnmarshalID(append(ZeroPublicKey[:], 1))
	assert.EqualError(t, err, io.ErrUnexpectedEOF.Error())

	_, err = UnmarshalID(append(ZeroPublicKey[:], append(net.IPv6loopback, 1)...))
	assert.EqualError(t, err, io.ErrUnexpectedEOF.Error())

	_, err = UnmarshalID(append(ZeroPublicKey[:], append(net.IPv6loopback, 1, 2)...))
	assert.NoError(t, err)
}
