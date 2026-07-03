package tree23

import (
	"math/rand"
	"testing"
)

var benchSizes = []int{1_000, 10_000, 100_000}

func buildTree(n int) (*Tree23[int], []int) {
	tr := New[int]()
	rng := rand.New(rand.NewSource(1))
	keys := make([]int, n)
	for i := 0; i < n; i++ {
		k := rng.Int()
		keys[i] = k
		tr.Insert(k)
	}
	return tr, keys
}

func BenchmarkInsert(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(sizeName(n), func(b *testing.B) {
			rng := rand.New(rand.NewSource(2))
			keys := make([]int, n)
			for i := range keys {
				keys[i] = rng.Int()
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tr := New[int]()
				for _, k := range keys {
					tr.Insert(k)
				}
			}
		})
	}
}

func BenchmarkSearch(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(sizeName(n), func(b *testing.B) {
			tr, keys := buildTree(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tr.Search(keys[i%len(keys)])
			}
		})
	}
}

func BenchmarkRangeQuery(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(sizeName(n), func(b *testing.B) {
			tr, _ := buildTree(n)
			const width = 1 << 20
			rng := rand.New(rand.NewSource(3))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lo := rng.Int()
				_ = tr.RangeQuery(lo, lo+width)
			}
		})
	}
}

func sizeName(n int) string {
	switch {
	case n >= 100_000:
		return "100k"
	case n >= 10_000:
		return "10k"
	default:
		return "1k"
	}
}
