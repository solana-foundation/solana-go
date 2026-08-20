package zk

import (
	"sync"
	"testing"
)

func TestConcurrentProofGeneration(t *testing.T) {
	kp := genKeyPair(t)
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			proof, err := PubkeyValidityProof(kp)
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
