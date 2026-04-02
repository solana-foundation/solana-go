package token2022

import (
	"context"
	"log"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	sendandconfirmtransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func TestCreateScaledUiAmountInstruction(t *testing.T) {
	t.Skip()
	rpcUrl, err := getRpcUrl()
	if err != nil {
		t.Fatal(err)
	}
	defer clearMirrorchain(rpcUrl)
	rpcClient := rpc.New(rpcUrl)

	payer, err := solana.NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rpcClient.RequestAirdrop(
		context.Background(),
		payer.PublicKey(),
		solana.LAMPORTS_PER_SOL,
		rpc.CommitmentConfirmed,
	); err != nil {
		t.Fatal(err)
	}

	mint, err := solana.NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	lamports, err := rpcClient.GetMinimumBalanceForRentExemption(context.Background(), DEFAULT_SCALED_UI_AMOUNT_MINT_LEN, rpc.CommitmentFinalized)
	if err != nil {
		t.Fatal(err)
	}

	createAccountIx := system.NewCreateAccountInstruction(
		lamports,
		DEFAULT_SCALED_UI_AMOUNT_MINT_LEN,
		ProgramID,
		payer.PublicKey(),
		mint.PublicKey(),
	)

	scaledUiAmuIx := NewScaledUiAmountInstruction(
		mint.PublicKey(),
		payer.PublicKey(),
		3.9,
	)
	token.SetProgramID(ProgramID)
	initializeMintIx := token.NewInitializeMintInstruction(
		2,
		payer.PublicKey(),
		payer.PublicKey(),
		mint.PublicKey(),
		solana.SysVarRentPubkey, //TODO: This is a really weird way of doing things... just declare it for me
	)

	recentBlockhash, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := solana.NewTransaction([]solana.Instruction{
		createAccountIx.Build(),
		scaledUiAmuIx,
		initializeMintIx.Build(),
	}, recentBlockhash.Value.Blockhash)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &payer
	}); err != nil {
		t.Fatal(err)
	}

	wsUrl := strings.Replace(rpcUrl, "https://", "wss://", 1)
	log.Println("wsUrl:", wsUrl)
	wsClient, err := ws.Connect(context.Background(), wsUrl)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sendandconfirmtransaction.SendAndConfirmTransaction(context.TODO(), rpcClient, wsClient, tx); err != nil {
		t.Fatal(err)
	}

}
