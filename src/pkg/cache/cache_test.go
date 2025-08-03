package cache_test

import (
	"flag"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/eldarbr/go-auth/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ = flag.String("t-db-uri", "", "perform sql tests on the `t-db-uri` database")

func TestMegaSetLowCap(t *testing.T) {
	t.Parallel()

	cache := cache.NewCache(1, 10)

	var waitGroup sync.WaitGroup

	for range 100 {
		waitGroup.Add(1)

		go func() {
			for range 10000 {
				// 100 threads are guaranteed to touch every random key withtin the expiration time.
				i := rand.Intn(15) //nolint:gosec // not a sensitive generation.
				key := strconv.Itoa(i)
				cache.Set(key, i)
			}

			waitGroup.Done()
		}()
	}

	waitGroup.Wait()

	time.Sleep(1 * time.Second)
	cache.DoAutoEvict()
}

func TestGetAndIncreaseParallel(t *testing.T) {
	t.Parallel()

	cache := cache.NewCache(2, 100000)

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

	time.Sleep(2 * time.Second)
	cache.DoAutoEvict() // manually clean up
}

func TestPeek(t *testing.T) {
	t.Parallel()

	tCache := cache.NewCache(2, 10)

	tCache.Set("1", 1)
	tCache.Set("2", 2)
	tCache.Set("3", 3)

	expTimer := time.NewTimer(2 * time.Second) // after the timer the keys will be guaranteed to be expired.

	time.Sleep(300 * time.Millisecond) // reasonable time to check if peek does not mess up with the expiration.

	val, err := tCache.Peek("1")
	require.NoError(t, err)
	assert.Equal(t, 1, val)

	val, err = tCache.Peek("2")
	require.NoError(t, err)
	assert.Equal(t, 2, val)

	val, err = tCache.Peek("3")
	require.NoError(t, err)
	assert.Equal(t, 3, val)

	_, err = tCache.Peek("4")
	require.ErrorIs(t, err, cache.ErrNoKey)

	<-expTimer.C // wait for the expiration

	_, err = tCache.Peek("1")
	require.ErrorIs(t, err, cache.ErrNoKey)

	_, err = tCache.Peek("2")
	require.ErrorIs(t, err, cache.ErrNoKey)

	_, err = tCache.Peek("3")
	require.ErrorIs(t, err, cache.ErrNoKey)

	_, err = tCache.Peek("4")
	require.ErrorIs(t, err, cache.ErrNoKey)
}

func TestGetExpired(t *testing.T) {
	t.Parallel()

	tCache := cache.NewCache(1, 10)

	tCache.Set("1", 1)
	tCache.Set("2", 2)
	tCache.Set("3", 3)

	time.Sleep(1000 * time.Millisecond) // after the timer the keys will be guaranteed to be expired.

	_, err := tCache.Get("1")
	require.ErrorIs(t, err, cache.ErrNoKey)

	_, err = tCache.Get("2")
	require.ErrorIs(t, err, cache.ErrNoKey)

	_, err = tCache.Get("3")
	require.ErrorIs(t, err, cache.ErrNoKey)

	_, err = tCache.Get("4")
	require.ErrorIs(t, err, cache.ErrNoKey)
}

func TestAutoEvict(t *testing.T) {
	t.Parallel()

	tCache := cache.NewCache(2, 10)

	go tCache.AutoEvict(1)

	tCache.Set("1", 1)
	tCache.Set("2", 2)
	tCache.Set("3", 3)

	time.Sleep(2000 * time.Millisecond) // after the timer the keys will be guaranteed to be expired.

	_, err := tCache.Get("1")
	require.ErrorIs(t, err, cache.ErrNoKey)

	_, err = tCache.Get("2")
	require.ErrorIs(t, err, cache.ErrNoKey)

	_, err = tCache.Get("3")
	require.ErrorIs(t, err, cache.ErrNoKey)

	_, err = tCache.Get("4")
	require.ErrorIs(t, err, cache.ErrNoKey)

	tCache.StopAutoEvict()
}
