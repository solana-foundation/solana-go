package zkprogram_test

import (
	"fmt"

	ag_solanago "github.com/solana-foundation/solana-go/v2"
	"github.com/solana-foundation/solana-go/v2/programs/system"
	zk_proofdata "github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/proofdata"
	zk_program "github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/zkprogram"
)

// Example_contextState records a verified statement in a proof context state
// account, then closes it to reclaim the rent.
//
// The program requires the account to be pre-allocated to exactly the size of
// the context it will hold, so the CreateAccount instruction has to run in
// the same transaction, ahead of the verification.
func Example_contextState() {
	var (
		payer        = ag_solanago.MustPublicKeyFromBase58("9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM")
		contextState = ag_solanago.MustPublicKeyFromBase58("11111111111111111111111111111112")
		authority    = payer
	)

	proof := new(zk_proofdata.PubkeyValidityProofData)

	space := zk_program.ContextStateSize(proof.ContextData())

	// rentExemptLamports would normally come from
	// rpc.Client.GetMinimumBalanceForRentExemption(ctx, space, commitment).
	const rentExemptLamports = 1_500_000

	create := system.NewCreateAccountInstruction(
		rentExemptLamports, space, zk_program.ProgramID, payer, contextState,
	).Build()

	info := &zk_program.ContextStateInfo{
		ContextStateAccount:   contextState,
		ContextStateAuthority: authority,
	}
	verify, err := zk_program.VerifyPubkeyValidity.EncodeVerifyProof(info, proof)
	if err != nil {
		panic(err)
	}

	// Once another program has read the statement, the authority signs a
	// close to recover the lamports.
	close := zk_program.CloseContextStateInstruction(*info, payer)

	verifyData, err := verify.Data()
	if err != nil {
		panic(err)
	}
	closeData, err := close.Data()
	if err != nil {
		panic(err)
	}

	fmt.Println("context state space:", space)
	fmt.Println("create accounts:", len(create.Accounts()))
	verifyType, _ := zk_program.InstructionType(verifyData)
	closeType, _ := zk_program.InstructionType(closeData)
	fmt.Println("verify instruction:", verifyType, "data:", len(verifyData), "bytes")
	fmt.Println("close instruction:", closeType, "data:", len(closeData), "bytes")
	fmt.Println("close authority signs:", close.Accounts()[2].IsSigner)

	// Output:
	// context state space: 65
	// create accounts: 2
	// verify instruction: VerifyPubkeyValidity data: 97 bytes
	// close instruction: CloseContextState data: 1 bytes
	// close authority signs: true
}
