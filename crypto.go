package valkyrie

import (
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type safeUlid struct {
	safe      *safeMonotonicReader
	t         time.Time
	monotonic *ulid.MonotonicEntropy
}

func (s *safeUlid) SafeMonotonic() ulid.ULID {
	return ulid.MustNew(ulid.Timestamp(s.t), s.safe)
}

func (s *safeUlid) Monotonic() ulid.ULID {
	return ulid.MustNew(ulid.Timestamp(s.t), s.monotonic)
}

// ULID Universally Unique Lexicographically Sortable Identifier
func ULID(t time.Time) *safeUlid {
	src := rand.NewSource(time.Now().UnixNano())
	entropy := rand.New(src)
	monotonic := ulid.Monotonic(entropy, 0)
	return &safeUlid{
		safe:      &safeMonotonicReader{MonotonicReader: monotonic},
		t:         t,
		monotonic: monotonic,
	}
}

type safeMonotonicReader struct {
	mtx sync.Mutex
	ulid.MonotonicReader
}

func (r *safeMonotonicReader) MonotonicRead(ms uint64, p []byte) (err error) {
	r.mtx.Lock()
	err = r.MonotonicReader.MonotonicRead(ms, p)
	r.mtx.Unlock()
	return err
}
