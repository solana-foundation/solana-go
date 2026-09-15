package zkprogram

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
)

const parityProofDataOffset = 0xDEADBEEF

var (
	parityContextStateAccount   = parityAddress(1)
	parityContextStateAuthority = parityAddress(2)
	parityProofAccount          = parityAddress(3)
	parityDestination           = parityAddress(4)
)

func parityAddress(b byte) solana.PublicKey {
	var key solana.PublicKey
	for i := range key {
		key[i] = b
	}
	return key
}

type parityAccount struct {
	Pubkey     string `json:"pubkey"`
	IsSigner   bool   `json:"is_signer"`
	IsWritable bool   `json:"is_writable"`
}

type parityInstruction struct {
	ProgramID string          `json:"program_id"`
	Data      string          `json:"data"`
	Accounts  []parityAccount `json:"accounts"`
}

type parityProof struct {
	Name          string            `json:"name"`
	ProofType     uint32            `json:"proof_type"`
	ProofData     string            `json:"proof_data"`
	Context       string            `json:"context"`
	ContextState  string            `json:"context_state"`
	VerifyNoCtx   parityInstruction `json:"verify_no_context"`
	VerifyWithCtx parityInstruction `json:"verify_with_context"`
	FromAcctNoCtx parityInstruction `json:"verify_from_account_no_context"`
	FromAcctCtx   parityInstruction `json:"verify_from_account_with_context"`
}

type parityData struct {
	ProgramID         string            `json:"program_id"`
	CloseContextState parityInstruction `json:"close_context_state"`
	Proofs            []parityProof     `json:"proofs"`
}

func TestParity(t *testing.T) {
	t.Parallel()
	parity_data := loadParityData(t)

	if !bytes.Equal(ProgramID.Bytes(), unhex(t, parity_data.ProgramID)) {
		t.Fatalf("program ID = %s, want %s", hex.EncodeToString(ProgramID.Bytes()), parity_data.ProgramID)
	}
	if len(parity_data.Proofs) != NumProofTypes {
		t.Fatalf("got %d proof vectors, want %d: every proof type needs a vector", len(parity_data.Proofs), NumProofTypes)
	}

	parityContextStateInfo := &ContextStateInfo{
		ContextStateAccount:   parityContextStateAccount,
		ContextStateAuthority: parityContextStateAuthority,
	}

	t.Run("CloseContextState", func(t *testing.T) {
		t.Parallel()
		checkInstructionMatches(t, parity_data.CloseContextState,
			CloseContextStateInstruction(*parityContextStateInfo, parityDestination))
	})

	for _, proofData := range parity_data.Proofs {
		t.Run(proofData.Name, func(t *testing.T) {
			checkProofDataMatches(t, proofData, parityContextStateInfo)
		})
	}
}

func loadParityData(t *testing.T) parityData {
	t.Helper()
	raw, err := os.ReadFile("testdata/rust_parity.json")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var parity_data parityData
	if err := json.Unmarshal(raw, &parity_data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	return parity_data
}

func checkProofDataMatches(t *testing.T, proofData parityProof, parityContextStateInfo *ContextStateInfo) {
	t.Parallel()
	typ := proofdata.ProofType(proofData.ProofType)
	if typ.String() != proofData.Name {
		t.Fatalf("proof type %d names itself %s, want %s", proofData.ProofType, typ, proofData.Name)
	}

	proof := fillProofData(t, typ)
	if !bytes.Equal(proof.Bytes(), unhex(t, proofData.ProofData)) {
		t.Errorf("proof data = %s, want %s", hex.EncodeToString(proof.Bytes()), proofData.ProofData)
	}

	context := proof.ContextData()
	if !bytes.Equal(context.Bytes(), unhex(t, proofData.Context)) {
		t.Errorf("context = %s, want %s", hex.EncodeToString(context.Bytes()), proofData.Context)
	}

	proofInstruction := ProofInstruction(typ)

	inline, err := proofInstruction.EncodeVerifyProof(nil, proof)
	if err != nil {
		t.Fatalf("EncodeVerifyProof: %v", err)
	}
	checkInstructionMatches(t, proofData.VerifyNoCtx, inline)

	inlineWithContext, err := proofInstruction.EncodeVerifyProof(parityContextStateInfo, proof)
	if err != nil {
		t.Fatalf("EncodeVerifyProof with context: %v", err)
	}
	checkInstructionMatches(t, proofData.VerifyWithCtx, inlineWithContext)

	fromAccount, err := proofInstruction.EncodeVerifyProofFromAccount(nil, parityProofAccount, 64)
	if err != nil {
		t.Fatalf("EncodeVerifyProofFromAccount: %v", err)
	}
	checkInstructionMatches(t, proofData.FromAcctNoCtx, fromAccount)

	fromAccountWithContext, err := proofInstruction.EncodeVerifyProofFromAccount(
		parityContextStateInfo, parityProofAccount, parityProofDataOffset)
	if err != nil {
		t.Fatalf("EncodeVerifyProofFromAccount with context: %v", err)
	}
	checkInstructionMatches(t, proofData.FromAcctCtx, fromAccountWithContext)

	// ProofContextState::encode, which is what the program writes to
	// the context state account.
	encoded, err := EncodeProofContextState(parityContextStateAuthority, typ, context)
	if err != nil {
		t.Fatalf("EncodeProofContextState: %v", err)
	}
	if !bytes.Equal(encoded, unhex(t, proofData.ContextState)) {
		t.Errorf("context state = %s, want %s", hex.EncodeToString(encoded), proofData.ContextState)
	}
	if got, want := len(encoded), int(ContextStateSize(context)); got != want {
		t.Errorf("encoded length = %d, want %d", got, want)
	}
}

// checkInstructionMatches compares a built instruction against the Rust encoding.
func checkInstructionMatches(t *testing.T, want parityInstruction, got solana.Instruction) {
	t.Helper()

	if !bytes.Equal(got.ProgramID().Bytes(), unhex(t, want.ProgramID)) {
		t.Errorf("program ID = %x, want %s", got.ProgramID().Bytes(), want.ProgramID)
	}

	data, err := got.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if !bytes.Equal(data, unhex(t, want.Data)) {
		t.Errorf("data = %s, want %s", hex.EncodeToString(data), want.Data)
	}

	accounts := got.Accounts()
	if len(accounts) != len(want.Accounts) {
		t.Fatalf("got %d accounts, want %d", len(accounts), len(want.Accounts))
	}
	for i, wantAccount := range want.Accounts {
		if !bytes.Equal(accounts[i].PublicKey.Bytes(), unhex(t, wantAccount.Pubkey)) {
			t.Errorf("account %d pubkey = %x, want %s", i, accounts[i].PublicKey.Bytes(), wantAccount.Pubkey)
		}
		if accounts[i].IsSigner != wantAccount.IsSigner {
			t.Errorf("account %d signer = %t, want %t", i, accounts[i].IsSigner, wantAccount.IsSigner)
		}
		if accounts[i].IsWritable != wantAccount.IsWritable {
			t.Errorf("account %d writable = %t, want %t", i, accounts[i].IsWritable, wantAccount.IsWritable)
		}
	}
}

func unhex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hex.DecodeString(%q): %v", s, err)
	}
	return b
}
