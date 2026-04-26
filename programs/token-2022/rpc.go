package token2022

import (
	"context"
	"fmt"

	bin "github.com/gagliardetto/solana-go/binary"
	"github.com/gagliardetto/solana-go/rpc"
)

const MINT_SIZE = 82

func (mint *Mint) Decode(data []byte) error {
	mint = new(Mint)
	dec := bin.NewBinDecoder(data)
	if err := dec.Decode(&mint); err != nil {
		return fmt.Errorf("unable to decode mint: %w", err)
	}
	return nil
}

func FetchMints(ctx context.Context, rpcCli *rpc.Client) (out []*Mint, err error) {
	resp, err := rpcCli.GetProgramAccountsWithOpts(
		ctx,
		ProgramID,
		&rpc.GetProgramAccountsOpts{
			Filters: []rpc.RPCFilter{
				{
					DataSize: MINT_SIZE,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("resp empty... program account not found")
	}

	for _, keyedAcct := range resp {
		acct := keyedAcct.Account

		m := new(Mint)
		if err := m.Decode(acct.Data.GetBinary()); err != nil {
			return nil, fmt.Errorf("unable to decode mint %q: %w", acct.Owner.String(), err)
		}
		out = append(out, m)

	}
	return
}
