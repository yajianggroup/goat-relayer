package bls

import (
	"context"
	"sync"
	"testing"
	"time"

	blst "github.com/supranational/blst/bindings/go"

	"github.com/goatnetwork/goat-relayer/internal/db"
	"github.com/goatnetwork/goat-relayer/internal/p2p"
	"github.com/stretchr/testify/assert"
)

var blsMode = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

type PrivateKey = blst.SecretKey
type PublicKey = blst.P2Affine
type Signature = blst.P1Affine

type AggregatePublicKey = blst.P2Aggregate
type AggregateSignature = blst.P1Aggregate

func TestSignatureHelperWithTwoNodes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// mock db
	dbm := db.NewDatabaseManager()

	// Create two LibP2PService instances
	p2psrv1 := p2p.NewLibP2PService(dbm)
	p2psrv2 := p2p.NewLibP2PService(dbm)

	// Start the two services
	go p2psrv1.Start(ctx)
	go p2psrv2.Start(ctx)

	// Wait for the services to start
	time.Sleep(2 * time.Second)

	// Create two SignatureHelper instances
	helper1, err := NewSignatureHelper(2)
	assert.NoError(t, err)

	helper2, err := NewSignatureHelper(2)
	assert.NoError(t, err)

	// Prepare the message to be signed
	message := []byte("Hello, world!")

	// Sign the message in parallel on both nodes
	var sig1, sig2 []byte
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		sig1 = helper1.SignDoc(ctx, message)
	}()

	go func() {
		defer wg.Done()
		sig2 = helper2.SignDoc(ctx, message)
	}()

	wg.Wait()

	// Verify the signature results
	assert.NotNil(t, sig1)
	assert.NotNil(t, sig2)
	assert.Equal(t, sig1, sig2)

	// Verify the aggregated signature
	pks := [][]byte{
		helper1.pk.Compress(),
		helper2.pk.Compress(),
	}

	isValid := AggregateVerify(pks, message, sig1)
	assert.True(t, isValid)
}

func AggregateVerify(pks [][]byte, msg, sig []byte) bool {
	if len(pks) == 0 {
		return false
	}

	signature := new(Signature).Uncompress(sig)
	if signature == nil {
		return false
	}

	pubkeys := make([]*PublicKey, 0, len(pks))
	for _, v := range pks {
		pk := new(PublicKey).Uncompress(v)
		if pk == nil {
			return false
		}
		pubkeys = append(pubkeys, pk)
	}
	return signature.FastAggregateVerify(true, pubkeys, msg, blsMode)
}
