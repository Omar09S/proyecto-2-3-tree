package tree23

import (
	"math/rand"
	"testing"
)

// Benchmarks del Entregable 2: miden empíricamente el comportamiento O(log n)
// de Insert y Search, y el O(log n + m) de RangeQuery.
//
// Ejecutar con:
//
//	go test ./tree23/ -bench=. -benchmem
//
// Para ver el crecimiento logarítmico, comparar el ns/op entre los tamaños
// 1e3, 1e4, 1e5: al multiplicar n por 10, el costo por operación sube solo una
// cantidad aproximadamente constante (un "escalón"), no x10.

var benchSizes = []int{1_000, 10_000, 100_000}

// buildTree construye un árbol con n claves aleatorias usando una semilla fija
// (resultados reproducibles entre corridas).
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

// BenchmarkInsert mide el costo de construir un árbol de n claves. Dividiendo
// el ns/op entre n se obtiene el costo promedio por inserción; si crece de forma
// aproximadamente logarítmica al pasar de 1k -> 10k -> 100k, se confirma O(log n)
// amortizado por inserción (O(n log n) la construcción completa).
func BenchmarkInsert(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(sizeName(n), func(b *testing.B) {
			// Pre-generamos las claves para no medir el costo del RNG.
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
				// Buscamos una clave existente (peor caso típico: llega a hoja).
				_ = tr.Search(keys[i%len(keys)])
			}
		})
	}
}

func BenchmarkRangeQuery(b *testing.B) {
	for _, n := range benchSizes {
		b.Run(sizeName(n), func(b *testing.B) {
			tr, _ := buildTree(n)
			// Rango acotado para que m (resultados) sea pequeño y se observe el
			// término O(log n) del costo.
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
