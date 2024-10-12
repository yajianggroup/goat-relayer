package http

import (
	"context"
	"encoding/hex"

	log "github.com/sirupsen/logrus"
)

type FireblocksProposal struct {
	BaseAPI           string
	Bip44AddressIndex uint32
	Bip44Change       uint32
	DerivationPath    [5]uint32
	VaultAccountId    string
	AssetId           string
}

func NewFireblocksProposal() *FireblocksProposal {
	// TODO: get base api from config network
	return &FireblocksProposal{BaseAPI: "", Bip44AddressIndex: 0, Bip44Change: 0, DerivationPath: [5]uint32{44, 0, 0, 0, 0}}
}

func (p *FireblocksProposal) PostRawSigningRequest(ctx context.Context, unsignedVouts [][]byte, note string) (*FbCreateTransactionResponse, error) {
	messages := make([]FbUnsignedRawMessage, len(unsignedVouts))
	for i, vout := range unsignedVouts {
		messages[i] = FbUnsignedRawMessage{Content: hex.EncodeToString(vout[:])}
	}
	request := &FbCreateTransactionRequest{
		Operation: "RAW",
		Source: FbSource{
			Type: "VAULT_ACCOUNT",
			ID:   p.VaultAccountId,
		},
		AssetID: p.AssetId,
		ExtraParameters: &FbExtraParameters{
			RawMessageData: FbRawMessageData{
				Messages: messages,
			},
		},
		Note: note,
	}

	log.Infof("PostRawSigningRequest to fireblocks: %v", request)

	//TODO: implement

	return nil, nil
}

func (p *FireblocksProposal) QueryTransaction(ctx context.Context, txid string) (*TransactionDetails, error) {
	// TODO: implement
	return nil, nil
}
