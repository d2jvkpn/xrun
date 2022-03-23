package internal

import (
	"sync"
	"testing"
)

func TestWG(t *testing.T) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	wg.Done()
	wg.Done()

	wg.Wait()
	wg.Wait()
}
