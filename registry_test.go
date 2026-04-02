package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterInstructionDecoder(t *testing.T) {

	decoder := func(instructionAccounts []*AccountMeta, data []byte) (any, error) {
		return nil, nil
	}
	decoderAnother := func(instructionAccounts []*AccountMeta, data []byte) (any, error) {
		return nil, nil
	}

	assert.NotPanics(t, func() {
		RegisterInstructionDecoder(BPFLoaderProgramID, decoder)
	})
	assert.NotPanics(t, func() {
		RegisterInstructionDecoder(BPFLoaderProgramID, decoder)
	})
	assert.Panics(t, func() {
		RegisterInstructionDecoder(BPFLoaderProgramID, decoderAnother)
	})
}
