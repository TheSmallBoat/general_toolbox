package core_module_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"awesomeProject/beacon/p2p_network/libs/cryptographic"

	"github.com/oasislabs/ed25519"
	"github.com/stretchr/testify/assert"
)

func TestMarshalJSON(t *testing.T) {
	pub, pri, err0 := ed25519.GenerateKey(nil)
	assert.NoError(t, err0)

	var pubKey cryptographic.PublicKey
	var priKey cryptographic.PrivateKey

	copy(pubKey[:], pub)
	copy(priKey[:], pri)

	pubKeyJSON, err1 := json.Marshal(pubKey)
	assert.NoError(t, err1)

	priKeyJSON, err2 := json.Marshal(priKey)
	assert.NoError(t, err2)

	assert.Equal(t, "\""+hex.EncodeToString(pub)+"\"", string(pubKeyJSON))
	assert.Equal(t, "\""+hex.EncodeToString(pri)+"\"", string(priKeyJSON))
}
