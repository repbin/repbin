package hashcash

import (
	"encoding/binary"
	"runtime"
)

var (
	// Steps is the number of steps per thread per cycle.
	Steps = uint64(4194304)
	// SingleThreadBits Number of bits required for parallel computing.
	SingleThreadBits = byte(20)
)

type computeResult struct {
	ok    bool
	nonce []byte
}

// ComputeNonceParallel computes hashcash in parallel.
func ComputeNonceParallel(d []byte, bits byte, start, stop uint64) (nonce []byte, ok bool) {
	ncpus := runtime.NumCPU()
	if ncpus > 1 {
		ncpus--
	}
	c := make(chan computeResult)
	prev := runtime.GOMAXPROCS(ncpus)
	defer runtime.GOMAXPROCS(prev)
	for {
		for i := ncpus; i != 0; i-- {
			go computeNonceThread(d, bits, start, start+Steps, c)
			start += Steps
		}
		for j := ncpus; j != 0; j-- {
			res := <-c
			if res.ok {
				return res.nonce, true
			}
		}
		if start >= stop && stop > 0 {
			ret := make([]byte, 8)
			binary.LittleEndian.PutUint64(ret, start)
			return ret, false
		}
	}
}

// computeNonceThread computes a nonce.
func computeNonceThread(d []byte, bits byte, c, stop uint64, res chan<- computeResult) {
	nonce, ok := ComputeNonce(d, bits, c, stop)
	res <- computeResult{
		nonce: nonce,
		ok:    ok,
	}
}

// ComputeNonceSelect calls parallel or single compute depending on bits.
func ComputeNonceSelect(d []byte, bits byte, start, stop uint64) (nonce []byte, ok bool) {
	if bits > SingleThreadBits {
		return ComputeNonceParallel(d, bits, start, stop)
	}
	return ComputeNonce(d, bits, start, stop)
}
