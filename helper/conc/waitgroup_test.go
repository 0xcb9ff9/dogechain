package conc

import (
	"context"
	"testing"

	"go.uber.org/atomic"

	"github.com/stretchr/testify/require"
)

func TestWaitGroup(t *testing.T) {
	t.Parallel()

	t.Run("all spawned run", func(t *testing.T) {
		t.Parallel()
		var count atomic.Int64
		wg := NewWaitGroup(context.TODO())

		for i := 0; i < 100; i++ {
			wg.Go(func(ctx context.Context) {
				count.Add(1)
			})
		}
		wg.Wait()
		require.Equal(t, count.Load(), int64(100))
	})

	t.Run("panic", func(t *testing.T) {
		t.Parallel()

		t.Run("is propagated", func(t *testing.T) {
			t.Parallel()
			wg := NewWaitGroup(context.TODO())
			wg.Go(func(ctx context.Context) {
				panic("super bad thing")
			})
			require.Panics(t, wg.Wait)
		})

		t.Run("one is propagated", func(t *testing.T) {
			t.Parallel()
			wg := NewWaitGroup(context.TODO())
			wg.Go(func(ctx context.Context) {
				panic("super bad thing")
			})
			wg.Go(func(ctx context.Context) {
				panic("super badder thing")
			})
			require.Panics(t, wg.Wait)
		})

		t.Run("non-panics do not overwrite panic", func(t *testing.T) {
			t.Parallel()
			wg := NewWaitGroup(context.TODO())
			wg.Go(func(ctx context.Context) {
				panic("super bad thing")
			})
			for i := 0; i < 10; i++ {
				wg.Go(func(ctx context.Context) {})
			}
			require.Panics(t, wg.Wait)
		})

		t.Run("non-panics run successfully", func(t *testing.T) {
			t.Parallel()
			wg := NewWaitGroup(context.TODO())
			var i atomic.Int64
			wg.Go(func(ctx context.Context) {
				i.Add(1)
			})
			wg.Go(func(ctx context.Context) {
				panic("super bad thing")
			})
			wg.Go(func(ctx context.Context) {
				i.Add(1)
			})
			require.Panics(t, wg.Wait)
			require.Equal(t, int64(2), i.Load())
		})
	})
}
