package bpfloader

import (
	"encoding/binary"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
)

const (
	PACKET_DATA_SIZE int = 1280 - 40 - 8
)

// https://github.com/solana-labs/solana/blob/v1.7.15/cli/src/program.rs#L1683
func calculateMaxChunkSize(
	createBuilder func(offset int, data []byte) *solana.TransactionBuilder,
) (size int, err error) {
	transaction, err := createBuilder(0, []byte{}).Build()
	if err != nil {
		return
	}
	signatures := make(
		[]solana.Signature,
		transaction.Message.Header.NumRequiredSignatures,
	)
	transaction.Signatures = append(transaction.Signatures, signatures...)
	serialized, err := transaction.MarshalBinary()
	if err != nil {
		return
	}
	size = PACKET_DATA_SIZE - len(serialized) - 1
	return
}

// https://github.com/solana-labs/solana/blob/v1.7.15/cli/src/program.rs#L2006
func completePartialProgramInit(
	loaderId solana.PublicKey,
	payerPubkey solana.PublicKey,
	elfPubkey solana.PublicKey,
	account *rpc.Account,
	accountDataLen int,
	minimumBalance uint64,
	allowExcessiveBalance bool,
) (instructions []solana.Instruction, balanceNeeded uint64, err error) {
	if account.Executable {
		err = fmt.Errorf("buffer account is already executable")
		return
	}
	if !account.Owner.Equals(loaderId) &&
		!account.Owner.Equals(solana.SystemProgramID) {
		err = fmt.Errorf(
			"buffer account passed is already in use by another program",
		)
		return
	}
	if len(account.Data.GetBinary()) > 0 &&
		len(account.Data.GetBinary()) < accountDataLen {
		err = fmt.Errorf(
			"buffer account passed is not large enough, may have been for a " +
				" different deploy?",
		)
		return
	}

	if len(account.Data.GetBinary()) == 0 &&
		account.Owner.Equals(solana.SystemProgramID) {
		instructions = append(
			instructions,
			system.NewAllocateInstruction(uint64(accountDataLen), elfPubkey).
				Build(),
		)
		instructions = append(
			instructions,
			system.NewAssignInstruction(loaderId, elfPubkey).Build(),
		)
		if account.Lamports < minimumBalance {
			balance := minimumBalance - account.Lamports
			instructions = append(
				instructions,
				system.NewTransferInstruction(balance, payerPubkey, elfPubkey).
					Build(),
			)
			balanceNeeded = balance
		} else if account.Lamports > minimumBalance &&
			account.Owner.Equals(solana.SystemProgramID) &&
			!allowExcessiveBalance {
			err = fmt.Errorf(
				"buffer account has a balance: %v.%v; it may already be in use",
				account.Lamports/solana.LAMPORTS_PER_SOL,
				account.Lamports%solana.LAMPORTS_PER_SOL,
			)
			return
		}
	}
	return
}

func load(
	payerPubkey solana.PublicKey,
	account *rpc.Account,
	programData []byte,
	bufferDataLen int,
	minimumBalance uint64,
	loaderId solana.PublicKey,
	bufferPubkey solana.PublicKey,
	allowExcessiveBalance bool,
) (
	initialBuilder *solana.TransactionBuilder,
	writeBuilders []*solana.TransactionBuilder,
	finalBuilder *solana.TransactionBuilder,
	balanceNeeded uint64,
	err error,
) {
	var instructions []solana.Instruction
	if account != nil {
		instructions, balanceNeeded, err = completePartialProgramInit(
			loaderId,
			payerPubkey,
			bufferPubkey,
			account,
			bufferDataLen,
			minimumBalance,
			allowExcessiveBalance,
		)
		if err != nil {
			return
		}
	} else {
		instructions = append(
			instructions,
			system.NewCreateAccountInstruction(
				minimumBalance,
				uint64(bufferDataLen),
				loaderId,
				payerPubkey,
				bufferPubkey,
			).Build(),
		)
		balanceNeeded = minimumBalance
	}
	if len(instructions) > 0 {
		initialBuilder = solana.NewTransactionBuilder().SetFeePayer(payerPubkey)
		for _, instruction := range instructions {
			initialBuilder = initialBuilder.AddInstruction(instruction)
		}
	}

	createBuilder := func(offset int, chunk []byte) *solana.TransactionBuilder {
		data := make([]byte, len(chunk)+16)
		binary.LittleEndian.PutUint32(data[0:], 0)
		binary.LittleEndian.PutUint32(data[4:], uint32(offset))
		binary.LittleEndian.PutUint32(data[8:], uint32(len(chunk)))
		binary.LittleEndian.PutUint32(data[12:], 0)
		copy(data[16:], chunk)
		instruction := solana.NewInstruction(
			loaderId,
			solana.AccountMetaSlice{
				solana.NewAccountMeta(bufferPubkey, true, true),
			},
			data,
		)
		return solana.NewTransactionBuilder().
			AddInstruction(instruction).
			SetFeePayer(payerPubkey)
	}

	chunkSize, err := calculateMaxChunkSize(createBuilder)
	if err != nil {
		return
	}
	writeBuilders = []*solana.TransactionBuilder{}
	for i := 0; i < len(programData); i += chunkSize {
		end := i + chunkSize
		if end > len(programData) {
			end = len(programData)
		}
		writeBuilders = append(
			writeBuilders,
			createBuilder(i, programData[i:end]),
		)
	}

	finalBuilder = solana.NewTransactionBuilder().SetFeePayer(payerPubkey)
	{
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data[0:], 1)
		instruction := solana.NewInstruction(
			loaderId,
			solana.AccountMetaSlice{
				solana.NewAccountMeta(bufferPubkey, true, true),
			},
			data,
		)
		finalBuilder.AddInstruction(instruction)
	}
	return
}

func Deploy(
	payerPubkey solana.PublicKey,
	account *rpc.Account,
	programData []byte,
	minimumBalance uint64,
	loaderId solana.PublicKey,
	bufferPubkey solana.PublicKey,
	allowExcessiveBalance bool,
) (
	initialBuilder *solana.TransactionBuilder,
	writeBuilders []*solana.TransactionBuilder,
	finalBuilder *solana.TransactionBuilder,
	balanceNeeded uint64,
	err error,
) {
	return load(
		payerPubkey,
		account,
		programData,
		len(programData),
		minimumBalance,
		loaderId,
		bufferPubkey,
		allowExcessiveBalance,
	)
}
