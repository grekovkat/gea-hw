package hw04lrucache

import (
	"math/rand"
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

	t.Run("purge logic", func(t *testing.T) {
		// (n = 3, добавили 4 элемента - 1й из кэша вытолкнулся);
		t.Log("purge by new")
		c := NewCache(3)

		wasInCache := c.Set("20", 20)
		require.False(t, wasInCache)

		wasInCache = c.Set("30", 30)
		require.False(t, wasInCache)

		wasInCache = c.Set("40", 40)
		require.False(t, wasInCache)

		wasInCache = c.Set("10", 50)
		require.False(t, wasInCache)

		v, isInCache := c.Get("20")
		require.Empty(t, v)
		require.False(t, isInCache)

		// проверяем, что вытолкнется тот, который наиболее давно был затронут.
		wasInCache = c.Set("30", 60)
		require.True(t, wasInCache)

		wasInCache = c.Set("40", 70)
		require.True(t, wasInCache)

		wasInCache = c.Set("10", 80)
		require.True(t, wasInCache)

		v, isInCache = c.Get("30")
		require.True(t, isInCache)
		require.Equal(t, 60, v)

		wasInCache = c.Set("50", 70)
		require.False(t, wasInCache)

		t.Log("purge oldest")
		v, isInCache = c.Get("40")
		require.Empty(t, v)
		require.False(t, isInCache)
	})
	t.Run("clear", func(t *testing.T) {
		// добавление элементов после очищения не приводит к панике
		c := NewCache(3)

		wasInCache := c.Set("A", 10)
		require.False(t, wasInCache)

		wasInCache = c.Set("B", 20)
		require.False(t, wasInCache)

		wasInCache = c.Set("C", 30)
		require.False(t, wasInCache)

		c.Clear()

		wasInCache = c.Set("D", 40)
		require.False(t, wasInCache)

		v, isInCache := c.Get("A")
		require.Empty(t, v)
		require.False(t, isInCache)

		v, isInCache = c.Get("D")
		require.Equal(t, 40, v)
		require.True(t, isInCache)
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
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
