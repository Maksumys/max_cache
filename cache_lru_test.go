package tools

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"runtime"
	"sync"
	"testing"
)

func TestList(t *testing.T) {
	list := NewList[string, int]()

	list.InsertTail("", 1)
	assert.Equal(t, list.Head.V, 1)
	assert.Nil(t, list.Head.Prev)
	assert.Nil(t, list.Head.Next)
	assert.Nil(t, list.Tail.Prev)
	assert.Nil(t, list.Tail.Next)
	assert.Equal(t, list.Tail.V, 1)

	list.InsertTail("", 2)
	assert.Equal(t, list.Head.V, 1)
	assert.Nil(t, list.Head.Prev)
	assert.NotNil(t, list.Head.Next)
	assert.Equal(t, list.Head.Next.V, 2)
	assert.NotNil(t, list.Tail)
	assert.Equal(t, list.Tail.V, 2)
	assert.NotNil(t, list.Tail.Prev)
	assert.Nil(t, list.Tail.Next)
	assert.Equal(t, list.Tail.Prev.V, 1)

	list.InsertTail("", 3)
	assert.Equal(t, list.Head.V, 1)
	assert.Nil(t, list.Head.Prev)
	assert.NotNil(t, list.Head.Next)
	assert.Equal(t, list.Head.Next.V, 2)
	assert.NotNil(t, list.Head.Next.Next)
	assert.Equal(t, list.Head.Next.Next.V, 3)

	assert.NotNil(t, list.Tail)
	assert.Equal(t, list.Tail.V, 3)
	assert.NotNil(t, list.Tail.Prev)
	assert.Nil(t, list.Tail.Next)
	assert.Equal(t, list.Tail.Prev.V, 2)

	assert.Equal(t, list.PopTail(), 3)
	assert.Equal(t, list.Head.V, 1)
	assert.Nil(t, list.Head.Prev)
	assert.NotNil(t, list.Head.Next)
	assert.Equal(t, list.Head.Next.V, 2)
	assert.NotNil(t, list.Tail)
	assert.Equal(t, list.Tail.V, 2)
	assert.NotNil(t, list.Tail.Prev)
	assert.Nil(t, list.Tail.Next)
	assert.Equal(t, list.Tail.Prev.V, 1)

	list.InsertFront("", 0)
	assert.Equal(t, list.Head.V, 0)
	assert.Nil(t, list.Head.Prev)
	assert.NotNil(t, list.Head.Next)
	assert.Equal(t, list.Head.Next.V, 1)
	assert.NotNil(t, list.Head.Next.Prev)
	assert.Equal(t, list.Head.Next.Prev.V, 0)
	assert.NotNil(t, list.Head.Next.Next)

	_, v := list.PopFront()

	assert.Equal(t, v, 0)
	assert.Equal(t, list.Head.V, 1)
	assert.Nil(t, list.Head.Prev)
	assert.NotNil(t, list.Head.Next)
	assert.Equal(t, list.Head.Next.V, 2)
	assert.NotNil(t, list.Tail)
	assert.Equal(t, list.Tail.V, 2)
	assert.NotNil(t, list.Tail.Prev)
	assert.Nil(t, list.Tail.Next)
	assert.Equal(t, list.Tail.Prev.V, 1)
}

func TestList_Reverse(t *testing.T) {
	list := NewList[string, int]()

	list.InsertTail("", 1)
	list.InsertTail("", 2)
	list.InsertTail("", 3)
	list.InsertTail("", 4)
	list.InsertTail("", 5)

	list.Reverse()

	fmt.Println(list)
}

func TestCache(t *testing.T) {
	cache := NewCacheLru[string, int](2, nil)
	assert.Equal(t, cache.length, 0)
	assert.Equal(t, cache.capacity, 2)

	cache.Set("q", 1)
	assert.Equal(t, cache.length, 1)

	v, ok := cache.Get("q")
	assert.True(t, ok)
	assert.Equal(t, v, 1)

	v, ok = cache.Get("w")
	assert.False(t, ok)
	v, ok = cache.Get("w")
	assert.False(t, ok)

	cache.Set("q", 2)
	assert.Equal(t, cache.length, 1)

	v, ok = cache.Get("q")
	assert.True(t, ok)
	assert.Equal(t, v, 2)

	cache.Set("q", 3)
	assert.Equal(t, cache.length, 1)

	v, ok = cache.Get("q")
	assert.True(t, ok)
	assert.Equal(t, v, 3)
}

func TestCacheSeveral(t *testing.T) {
	cache := NewCacheLru[string, int](2, nil)

	cache.Set("q", 1)
	assert.Equal(t, cache.length, 1)
	cache.Set("w", 2)
	assert.Equal(t, cache.length, 2)

	v, ok := cache.Get("q")
	assert.True(t, ok)
	assert.Equal(t, v, 1)
	v, ok = cache.Get("w")
	assert.True(t, ok)
	assert.Equal(t, v, 2)

	cache.Set("e", 3)
	assert.Equal(t, cache.length, 2)

	v, ok = cache.Get("q")
	assert.False(t, ok)

	v, ok = cache.Get("w")
	assert.True(t, ok)
	assert.Equal(t, v, 2)

	v, ok = cache.Get("e")
	assert.True(t, ok)
	assert.Equal(t, v, 3)

}

func TestCache_Erase1(t *testing.T) {
	cache := NewCacheLruLock[*sync.Mutex, string, int](4, nil, &sync.Mutex{})
	assert.NotPanics(t, func() {
		cache.Erase("1")
	})
	cache.Set("1", 1)
	assert.Equal(t, cache.cache.length, 1)

	v, ok := cache.Get("1")
	assert.Equal(t, v, 1)
	assert.True(t, ok)

	cache.Erase("1")

	assert.Equal(t, cache.cache.length, 0)

	v, ok = cache.Get("1")
	assert.False(t, ok)

	cache.Set("1", 1)
	cache.Set("2", 2)

	assert.Equal(t, cache.cache.length, 2)

	cache.Erase("1")
	assert.Equal(t, cache.cache.length, 1)

	v, ok = cache.Get("2")
	assert.Equal(t, v, 2)
	assert.True(t, ok)

	v, ok = cache.Get("1")
	assert.False(t, ok)

	cache.Erase("2")
	assert.Equal(t, cache.cache.length, 0)

	v, ok = cache.Get("2")
	assert.False(t, ok)
}

func TestCache_Erase2(t *testing.T) {
	cache := NewCacheLruLock[*sync.Mutex, string, int](4, nil, &sync.Mutex{})

	cache.Set("1", 1)
	cache.Set("2", 2)
	cache.Set("3", 3)

	cache.Erase("2")
	v, ok := cache.Get("1")
	assert.Equal(t, v, 1)
	assert.True(t, ok)
	v, ok = cache.Get("3")
	assert.Equal(t, v, 3)
	assert.True(t, ok)

	v, ok = cache.Get("2")
	assert.False(t, ok)
}

func TestCache_Erase3(t *testing.T) {
	cache := NewCacheLruLock[*sync.Mutex, string, int](4, nil, &sync.Mutex{})

	cache.Set("1", 1)
	cache.Set("2", 2)
	cache.Set("3", 3)
	cache.Set("4", 4)

	cache.Erase("2")
	v, ok := cache.Get("1")
	assert.Equal(t, v, 1)
	assert.True(t, ok)
	v, ok = cache.Get("3")
	assert.Equal(t, v, 3)
	assert.True(t, ok)
	v, ok = cache.Get("4")
	assert.Equal(t, v, 4)
	assert.True(t, ok)

	v, ok = cache.Get("2")
	assert.False(t, ok)
}

func TestCache_OnEvicted(t *testing.T) {
	i := 1
	cache := NewCacheLruLock[*sync.Mutex, string, int](4, func(v int) {
		assert.Equal(t, v, i)
		i++
	}, &sync.Mutex{})

	cache.Set("1", 1)
	cache.Set("2", 2)
	cache.Set("3", 3)
	cache.Set("4", 4)
	cache.Set("5", 5)
}

func BenchmarkInsert(b *testing.B) {
	cache := NewCacheLruLock[*sync.Mutex, string, int](4, func(v int) {
		assert.Equal(b, v, 1)
	}, &sync.Mutex{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cache.Set("q", 1)
	}
}

func BenchmarkGet(b *testing.B) {
	cache := NewCacheLruLock[*sync.Mutex, string, int](4, func(v int) {
		assert.Equal(b, v, 1)
	}, &sync.Mutex{})

	cache.Set("q", 1)
	cache.Set("w", 1)
	cache.Set("e", 1)
	cache.Set("q", 1)
	cache.Set("w", 1)
	cache.Set("e", 1)
	cache.Set("z", 1)
	cache.Set("s", 1)
	cache.Set("f", 1)
	cache.Set("g", 1)
	cache.Set("q", 1)
	cache.Set("qsdf", 1)
	cache.Set("qc", 1)
	cache.Set("q", 1)
	cache.Set("qv", 1)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get("q")
		}
	})
}

func benchmarkCache(b *testing.B, localWork, writeRatio int) {
	cache := NewCacheLruLock[*sync.Mutex, string, int](4, func(v int) {
		assert.Equal(b, v, 1)
	}, &sync.Mutex{})

	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			foo++
			if foo%writeRatio == 0 {
				cache.Set("q", 1)
			} else {
				cache.Get("q")
				for i := 0; i != localWork; i += 1 {
					foo *= 2
					foo /= 2
				}
			}
		}
		_ = foo
	})
}

func BenchmarkCacheWrite100(b *testing.B) {
	benchmarkCache(b, 0, 100)
}

func BenchmarkCacheWrite10(b *testing.B) {
	benchmarkCache(b, 0, 10)
}

func BenchmarkCacheWorkWrite100(b *testing.B) {
	benchmarkCache(b, 100, 100)
}

func BenchmarkCacheWorkWrite10(b *testing.B) {
	benchmarkCache(b, 100, 10)
}

func HammerMutex(m *SpinMutex, loops int, cdone chan bool) {
	for i := 0; i < loops; i++ {
		if i%3 == 0 {
			if m.TryLock() {
				m.Unlock()
			}
			continue
		}
		m.Lock()
		m.Unlock()
	}
	cdone <- true
}

func TestMutex(t *testing.T) {
	if n := runtime.SetMutexProfileFraction(1); n != 0 {
		t.Logf("got mutexrate %d expected 0", n)
	}
	defer runtime.SetMutexProfileFraction(0)

	m := new(SpinMutex)

	m.Lock()
	if m.TryLock() {
		t.Fatalf("TryLock succeeded with mutex locked")
	}
	m.Unlock()
	if !m.TryLock() {
		t.Fatalf("TryLock failed with mutex unlocked")
	}
	m.Unlock()

	c := make(chan bool)
	for i := 0; i < 10; i++ {
		go HammerMutex(m, 1000, c)
	}
	for i := 0; i < 10; i++ {
		<-c
	}
}

func benchmarkMutex(b *testing.B, slack, work bool) {
	var mu SpinMutex
	if slack {
		b.SetParallelism(10)
	}
	b.RunParallel(func(pb *testing.PB) {
		foo := 0
		for pb.Next() {
			mu.Lock()
			mu.Unlock()
			if work {
				for i := 0; i < 100; i++ {
					foo *= 2
					foo /= 2
				}
			}
		}
		_ = foo
	})
}

func BenchmarkMutex(b *testing.B) {
	benchmarkMutex(b, false, false)
}

func BenchmarkMutexSlack(b *testing.B) {
	benchmarkMutex(b, true, false)
}

func BenchmarkMutexWork(b *testing.B) {
	benchmarkMutex(b, false, true)
}

func BenchmarkMutexWorkSlack(b *testing.B) {
	benchmarkMutex(b, true, true)
}
