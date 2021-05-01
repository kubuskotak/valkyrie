package valkyrie

import (
	"fmt"
	"testing"
	"time"
)

func TestMonotonic(t *testing.T) {
	t.Parallel()

	t0 := time.Now()
	s := ULID(t0)
	errs := make(chan error, 100)
	for i := 0; i < cap(errs); i++ {
		u0 := s.Monotonic()
		u1 := s.Monotonic()
		for j := 0; j < 1024; j++ {
			u0, u1 = u1, s.Monotonic()
			if u0.String() >= u1.String() {
				errs <- fmt.Errorf(
					"%s (%d %x) >= %s (%d %x)",
					u0.String(), u0.Time(), u0.Entropy(),
					u1.String(), u1.Time(), u1.Entropy(),
				)
				return
			}
		}
		errs <- nil
	}
	for i := 0; i < cap(errs); i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
}

func TestMonotonicSafe(t *testing.T) {
	t.Parallel()

	t0 := time.Now()
	s := ULID(t0)
	errs := make(chan error, 100)
	for i := 0; i < cap(errs); i++ {
		go func() {
			u0 := s.SafeMonotonic()
			u1 := s.SafeMonotonic()
			for j := 0; j < 1024; j++ {
				u0, u1 = u1, s.SafeMonotonic()
				if u0.String() >= u1.String() {
					errs <- fmt.Errorf(
						"%s (%d %x) >= %s (%d %x)",
						u0.String(), u0.Time(), u0.Entropy(),
						u1.String(), u1.Time(), u1.Entropy(),
					)
					return
				}
			}
			errs <- nil
		}()
	}
	for i := 0; i < cap(errs); i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
}
