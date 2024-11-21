package cache_test

import (
	"flag"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/eldarbr/go-auth/internal/service/cache"
	"github.com/stretchr/testify/require"
	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
)

var _ = flag.String("t-db-uri", "", "perform sql tests on the `t-db-uri` database")

func TestMegaSetLowCap(t *testing.T) {
	t.Parallel()

	cache := cache.NewCache(999, 10)

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)

		go func() {
			for range 10000 {
				i := rand.Intn(15)
				key := strconv.Itoa(i)
				cache.Set(key, i)
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

func TestGetAndIncreaseParallel(t *testing.T) {
	t.Parallel()

	cache := cache.NewCache(999, 100000)

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)

		go func() {
			for range 10000 {
				cache.GetAndIncrease("123")
			}

			wg.Done()
		}()
	}

	wg.Wait()

	val, err := cache.Get("123")
	require.NoError(t, err)
	require.Equal(t, 100*10000, val)
}
