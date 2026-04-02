package token2022

import (
	"context"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	sendandconfirmtransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func TestCreateInitializeMetadataPointerInstruction(t *testing.T) {
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

	updateAuth, _ := solana.NewRandomPrivateKey()

	metadata := TokenMetadata{
		UpdateAuthority: updateAuth.PublicKey().ToPointer(),
		Mint:            mint.PublicKey(),
		Name:            "OPOS",
		Symbol:          "OPOS",
		Uri:             "https://raw.githubusercontent.com/solana-developers/opos-asset/main/assets/DeveloperPortal/metadata.json",
		AdditionalMetadata: map[string]string{
			"description": "Only Possible On Solana",
		},
	}

	lamports, err := rpcClient.GetMinimumBalanceForRentExemption(context.Background(), metadata.LenForLamports(), rpc.CommitmentFinalized)
	if err != nil {
		t.Fatal(err)
	}

	createAccountIx := system.NewCreateAccountInstruction(
		lamports,
		DEFAULT_METADATA_MINT_LEN,
		ProgramID,
		payer.PublicKey(),
		mint.PublicKey(),
	)

	initializeMetadataIx := NewInitializeMetadataPointerInstruction(
		mint.PublicKey(),
		payer.PublicKey(),
		mint.PublicKey(),
	)

	token.SetProgramID(ProgramID)
	initializeMintIx := token.NewInitializeMintInstructionBuilder().
		SetDecimals(2).
		SetMintAuthority(payer.PublicKey()).
		SetMintAccount(mint.PublicKey()).
		SetSysVarRentPubkeyAccount(solana.SysVarRentPubkey)

	initMetadataIx := CreateInitializeInstruction(InitializeInstructionArgs{
		Metadata:        mint.PublicKey(),
		UpdateAuthority: updateAuth.PublicKey(),
		Mint:            mint.PublicKey(),
		MintAuthority:   payer.PublicKey(),
		Name:            metadata.Name,
		Symbol:          metadata.Symbol,
		Uri:             metadata.Uri,
	})

	recentBlockhash, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := solana.NewTransaction([]solana.Instruction{
		createAccountIx.Build(),
		initializeMetadataIx,
		initializeMintIx.Build(),
		initMetadataIx,
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
	wsClient, err := ws.Connect(context.Background(), wsUrl)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sendandconfirmtransaction.SendAndConfirmTransaction(context.TODO(), rpcClient, wsClient, tx); err != nil {
		t.Fatal(err)
	}

	mintAccount, err := rpcClient.GetAccountInfo(context.Background(), mint.PublicKey())
	if err != nil {
		t.Fatal(err)
	}
	parsedMeta, err := ParseTokenMetadata(mintAccount.Value.Data.GetBinary())
	if err != nil {
		t.Fatal(err)
	}
	if *parsedMeta.UpdateAuthority != updateAuth.PublicKey() {
		t.Fatalf("UpdateAuthority mismatch")
	}
	if parsedMeta.Mint != mint.PublicKey() {
		t.Fatalf("Mint mismatch")
	}
	if parsedMeta.Name != metadata.Name {
		t.Fatalf("Name mismatch")
	}
	if parsedMeta.Symbol != metadata.Symbol {
		t.Fatalf("Symbol mismatch")
	}
	if parsedMeta.Uri != metadata.Uri {
		t.Fatalf("Uri mismatch")
	}
}
