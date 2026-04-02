package token2022

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	sendandconfirmtransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/joho/godotenv"
)

func TestCreateInitializeDefaultAccountStateInstruction(t *testing.T) {
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

	lamports, err := rpcClient.GetMinimumBalanceForRentExemption(context.Background(), DEFAULT_ACCOUNT_STATE_MINT_LEN, rpc.CommitmentFinalized)
	if err != nil {
		t.Fatal(err)
	}

	createAccountIx := system.NewCreateAccountInstruction(
		lamports,
		DEFAULT_ACCOUNT_STATE_MINT_LEN,
		ProgramID,
		payer.PublicKey(),
		mint.PublicKey(),
	)

	initializeDefaultAccountStateIx := NewInitializeDefaultAccountStateInstruction(
		mint.PublicKey(),
		token.Frozen,
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
		initializeDefaultAccountStateIx,
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

	updateIx := CreateUpdateDefaultAccountStateInstruction(
		mint.PublicKey(),
		token.Initialized,
		payer.PublicKey(),
	)

	recentBlockhash, err = rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		t.Fatal(err)
	}

	tx, err = solana.NewTransaction([]solana.Instruction{updateIx}, recentBlockhash.Value.Blockhash)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &payer
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := sendandconfirmtransaction.SendAndConfirmTransaction(context.TODO(), rpcClient, wsClient, tx); err != nil {
		t.Fatal(err)
	}

}

func getRpcUrl() (string, error) {
	godotenv.Load("../../.env")
	mirrorApiKey := os.Getenv("MIRROR_API_KEY")

	req, err := http.NewRequest("POST", "https://rpc.mirror.ad/blockchains", bytes.NewBuffer([]byte{}))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}
	req.Header.Set("api_key", mirrorApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}
	if resp.StatusCode != 200 {
		fmt.Println("Error getting RPC URL:", resp.Status)
		return "", fmt.Errorf("Error getting RPC URL: %s", resp.Status)
	}

	type res struct {
		Url string `json:"url"`
	}
	var r res
	err = json.Unmarshal(body, &r)
	if err != nil {
		fmt.Println("Error unmarshalling response body:", err)
		return "", err
	}
	return r.Url, nil

}

func clearMirrorchain(rpcUrl string) error {
	godotenv.Load("../../.env")
	mockchainApiKey := os.Getenv("MIRROR_API_KEY")
	req, err := http.NewRequest("DELETE", rpcUrl, nil)
	if err != nil {
		fmt.Println("Error creating request to clear mockchain:", err)
		return err
	}
	req.Header.Set("api_key", mockchainApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}
	if resp.StatusCode != 200 {
		fmt.Println("Error clearing mockchain:", resp.Status)
		return fmt.Errorf("Error clearing mockchain: %s", string(body))
	}
	return nil
}
