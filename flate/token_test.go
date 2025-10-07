package flate

import (
	"runtime"
	"sync"
	"testing"
)

type chTookensPool struct {
	ch chan *tokens
}

func (p *chTookensPool) Get() *tokens {
	select {
	case b := <-p.ch:
		return b
	default:
		return &tokens{}
	}
}

func (p *chTookensPool) Put(t *tokens) {
	select {
	case p.ch <- t: // ok
	default: // drop
	}
}

func NewTokensPool(max int) *chTookensPool {
	c := make(chan *tokens, max)
	for i := 0; i < max; i++ {
		c <- &tokens{}
	}
	return &chTookensPool{ch: c}
}

var tp = NewTokensPool(runtime.NumCPU())

func BenchmarkTokensInitialization(b *testing.B) {
	// BenchmarkTokensInitialization/stack-allocation-12         	   31321	     44704 ns/op	  270338 B/op	       1 allocs/op
	b.Run("stack-allocation", func(b *testing.B) {
		b.RunParallel(func(b *testing.PB) {
			for b.Next() {
				var dst tokens
				_ = dst
			}
		})
	})

	// BenchmarkTokensInitialization/pool-allocation-12          	808984578	         1.522 ns/op	       0 B/op	       0 allocs/op
	b.Run("pool-allocation", func(b *testing.B) {
		b.RunParallel(func(b *testing.PB) {
			pool := sync.Pool{
				New: func() interface{} { return &tokens{} },
			}
			for b.Next() {
				t := pool.Get().(*tokens)
				pool.Put(t)

			}
		})
	})

	// Probably better on low RPS (?)
	// BenchmarkTokensInitialization/chanels-allocation-12       	 9181482	       123.3 ns/op	       0 B/op	       0 allocs/op
	b.Run("chanels-allocation", func(b *testing.B) {
		b.RunParallel(func(b *testing.PB) {
			for b.Next() {
				t := tp.Get()
				tp.Put(t)

			}
		})
	})
}
