package rpc

import (
	"bytes"
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
)

// Metaplex Token Metadata
const (
	metaplex = "metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s"
	meta     = "metadata"
)

func (cl *Client) GetTokenMetaData(mint string) (*TokenMetaData, error) {
	b, err := base58.Decode(metaplex)
	if err != nil {
		return nil, err
	}
	seeds := [][]byte{
		[]byte(meta),
		b,
		solana.MustPublicKeyFromBase58(mint).Bytes(),
	}
	pb, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(metaplex))
	if err != nil {
		return nil, err
	}
	resp, err := cl.GetAccountInfoWithOpts(
		context.TODO(),
		pb,
		&GetAccountInfoOpts{
			Encoding:   solana.EncodingBase64,
			Commitment: CommitmentFinalized,
		},
	)
	if err != nil {
		return nil, err
	}
	metadata := &TokenMetaData{}
	data := resp.Value.Data.GetBinary()
	metadata.Name = string(bytes.Trim(data[69:69+32], "\x00"))
	metadata.Symbol = string(bytes.Trim(data[105:115], "\x00"))
	metadata.URI = string(bytes.Trim(data[116:200], "\x00"))

	return metadata, err
}
