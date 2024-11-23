package cache_test

import (
	"flag"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/eldarbr/go-auth/pkg/cache"
	"github.com/stretchr/testify/require"
)

var _ = flag.String("t-db-uri", "", "perform sql tests on the `t-db-uri` database")

func TestMegaSetLowCap(t *testing.T) {
	t.Parallel()

	cache := cache.NewCache(999, 10)

	var waitGroup sync.WaitGroup

	for range 100 {
		waitGroup.Add(1)

		go func() {
			for range 10000 {
				i := rand.Intn(15) //nolint:gosec // not a sensitive generation.
				key := strconv.Itoa(i)
				cache.Set(key, i)
			}

			waitGroup.Done()
		}()
	}

	waitGroup.Wait()
}

func TestGetAndIncreaseParallel(t *testing.T) {
	t.Parallel()

	cache := cache.NewCache(999, 100000)

	var waitGroup sync.WaitGroup

	for range 100 {
		waitGroup.Add(1)

		go func() {
			for range 10000 {
				cache.GetAndIncrease("123")
			}

			waitGroup.Done()
		}()
	}

	waitGroup.Wait()

	val, err := cache.Get("123")
	require.NoError(t, err)
	require.Equal(t, 100*10000, val)
}
