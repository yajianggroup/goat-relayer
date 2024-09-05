package tss

import (
	"context"

	"github.com/bnb-chain/tss-lib/v2/common"
	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	tsslib "github.com/bnb-chain/tss-lib/v2/tss"
)

type TSSService interface {
	HandleKeygenMessages(ctx context.Context, tssKeyInCh chan KeygenMessage, tssKeyOutCh chan tsslib.Message, tssKeyEndCh chan *keygen.LocalPartySaveData)
	HandleSigningMessages(ctx context.Context, tssSignInCh chan SigningMessage, tssSignOutCh chan tsslib.Message, tssSignEndCh chan *common.SignatureData)
}

type TSSServiceImpl struct{}

type KeygenMessage struct {
	Content string
}

type SigningMessage struct {
	Content string
}
