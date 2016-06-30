package rumutex

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestRUMutex(t *testing.T) {
	var (
		data int
		ru   RUMutex
		wg   sync.WaitGroup
	)
	const (
		N = 100
		C = 10
	)
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			switch i {
			case 0, N / 2:
				// Test to upgrade, then downgrade
				for c := 0; c < C; c++ {
					ru.RLock()
					d := data
					for !ru.Upgrade() {
						ru.RUnlock()
						time.Sleep(10 * time.Millisecond)
						ru.RLock()
						d = data
					}
					next := d + 7
					data = next
					ru.Downgrade()
					if data != next {
						t.Fatalf("data = %d, want %d", data, next)
					}
					ru.RUnlock()
				}
				wg.Done()
			case 1, N/2 + 1:
				// Test to upgrade, then unlock
				for c := 0; c < C; c++ {
					ru.RLock()
					d := data
					for !ru.Upgrade() {
						ru.RUnlock()
						runtime.Gosched()
						ru.RLock()
						d = data
					}
					data = d + 997
					ru.Unlock()
				}
				wg.Done()
			case 2, N/2 + 2:
				// Test to re-upgrade
				for c := 0; c < C; c++ {
					ru.RLock()
					d := data
					for !ru.Upgrade() {
						ru.RUnlock()
						runtime.Gosched()
						ru.RLock()
						d = data
					}
					next := d + 10007
					data = next
					ru.Downgrade()
					runtime.Gosched()
					if data != next {
						t.Fatalf("data = %d, want %d", data, next)
					}
					if ru.Upgrade() {
						if data != next {
							t.Fatalf("data = %d, want %d", data, next)
						}
						runtime.Gosched()
						ru.Unlock()
					} else {
						ru.RUnlock()
					}
				}
				wg.Done()
			default:
				// Test read lock
				for c := 0; c < 10; c++ {
					ru.RLock()
					time.Sleep(10 * time.Millisecond)
					ru.RUnlock()
				}
				wg.Done()
			}
		}(i)
	}
	wg.Wait()
	want := C*2*7 + C*2*997 + C*2*10007
	if data != want {
		t.Fatalf("data = %d, want %d", data, want)
	}
}
