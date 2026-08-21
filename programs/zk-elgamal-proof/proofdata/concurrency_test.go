package proofdata

import (
	"sync"
	"testing"

	"github.com/gagliardetto/solana-go/programs/zk-elgamal-proof/internal/zktest"
)

func TestConcurrentProofGeneration(t *testing.T) {
	kp := zktest.GenKeyPair(t)
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			proof, err := NewPubkeyValidityProofData(kp)
			if err != nil {
				errs <- err
				return
			}
			errs <- proof.Verify()
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}
