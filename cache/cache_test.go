package cache

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("extraction", func(t *testing.T) {
		c := NewCache(3)
		c.Set("1", 1)
		c.Set("2", 2)
		c.Set("3", 3)

		for i, key := range []Key{"1", "2", "3"} {
			item, ok := c.Get(key)
			require.True(t, ok)
			require.Equal(t, i+1, item)
		}

		// Вытеснение
		c.Set("4", 4)
		item, ok := c.Get("4")
		require.True(t, ok)
		require.Equal(t, 4, item)

		item, ok = c.Get("1")
		require.False(t, ok)
		require.Nil(t, item)
	})

	t.Run("extraction_oldest", func(t *testing.T) {
		c := NewCache(3)
		c.Set("1", 1)
		c.Set("2", 2)
		c.Set("3", 3)

		for i, key := range []Key{"1", "2", "3"} {
			item, ok := c.Get(key)
			require.True(t, ok)
			require.Equal(t, i+1, item)
		} // [3 2 1]

		c.Get("2") // [2 3 1]
		c.Get("3") // [3 2 1]

		// Вытеснение
		c.Set("4", 4) // [4 3 2]
		item, ok := c.Get("1")
		t.Log(item, ok)
		require.False(t, ok)
		require.Nil(t, item)
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000)))) //nolint:gosec
		}
	}()

	wg.Wait()
}

func TestCacheWithCustomDeleter(t *testing.T) {
	/*
		Тест для проверки удаления файлов с диска
	*/
	files := make([]*os.File, 0)
	require.NoError(t, os.MkdirAll("tmp", 0o755))
	defer func() {
		require.NoError(t, os.RemoveAll("tmp"))
		for _, f := range files {
			require.NoError(t, f.Close())
		}
		ResetDeleter()
	}()

	c := NewCache(2)
	NewCustomDeleter(func(value interface{}) {
		require.NoError(t, os.RemoveAll(value.(string)))
	})

	for i := 0; i < 5; i++ {
		f, err := os.CreateTemp("tmp", "file_")
		require.NoError(t, err)
		files = append(files, f)
		c.Set(Key(f.Name()), f.Name())
		entries, err := os.ReadDir("tmp")
		require.NoError(t, err)
		require.LessOrEqual(t, len(entries), 2)
	}
}
