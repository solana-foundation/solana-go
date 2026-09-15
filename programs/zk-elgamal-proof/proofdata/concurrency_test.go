package proofdata

import (
	"sync"
	"testing"

	"github.com/solana-foundation/solana-go/v2/programs/zk-elgamal-proof/internal/zktest"
)

func TestConcurrentProofGeneration(t *testing.T) {
	const N = 200
	kp := zktest.GenKeyPair(t)
	var wg sync.WaitGroup
	errs := make(chan error, N)
	for i := 0; i < N; i++ {
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
