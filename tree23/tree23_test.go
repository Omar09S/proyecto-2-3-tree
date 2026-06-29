package tree23

import (
	"math/rand"
	"sort"
	"testing"
)

// Pruebas del Paso 1: validan la navegación de Search sobre árboles
// construidos manualmente, ANTES de implementar la inserción automática.
// Esto nos permite confiar en que el descenso por el árbol es correcto.

// buildManualTree construye a mano el siguiente árbol 2-3 de altura 2 para
// probar la búsqueda sin depender todavía de Insert:
//
//	             [ 10 | 20 ]            (raíz: 3-nodo)
//	            /     |     \
//	      [5]      [15]      [25]       (hojas: 2-nodos)
//
// Claves presentes: 5, 10, 15, 20, 25.
func buildManualTree() *Tree23[int] {
	left := &Node[int]{numKeys: 1}
	left.keys[0] = 5

	middle := &Node[int]{numKeys: 1}
	middle.keys[0] = 15

	right := &Node[int]{numKeys: 1}
	right.keys[0] = 25

	root := &Node[int]{numKeys: 2}
	root.keys[0] = 10
	root.keys[1] = 20
	root.children[0] = left
	root.children[1] = middle
	root.children[2] = right

	return &Tree23[int]{root: root, size: 5}
}

func TestSearchEmptyTree(t *testing.T) {
	tr := New[int]()
	if tr.Search(42) {
		t.Errorf("Search(42) en árbol vacío = true; se esperaba false")
	}
	if tr.Len() != 0 {
		t.Errorf("Len() = %d en árbol vacío; se esperaba 0", tr.Len())
	}
}

func TestSearchSingleKey(t *testing.T) {
	root := &Node[int]{numKeys: 1}
	root.keys[0] = 7
	tr := &Tree23[int]{root: root, size: 1}

	if !tr.Search(7) {
		t.Errorf("Search(7) = false; se esperaba true (clave presente)")
	}
	if tr.Search(3) {
		t.Errorf("Search(3) = true; se esperaba false (clave ausente)")
	}
}

func TestSearchPresentKeys(t *testing.T) {
	tr := buildManualTree()
	for _, k := range []int{5, 10, 15, 20, 25} {
		if !tr.Search(k) {
			t.Errorf("Search(%d) = false; la clave debería estar presente", k)
		}
	}
}

func TestSearchAbsentKeys(t *testing.T) {
	tr := buildManualTree()
	for _, k := range []int{0, 6, 12, 18, 22, 30} {
		if tr.Search(k) {
			t.Errorf("Search(%d) = true; la clave NO debería estar presente", k)
		}
	}
}

func TestSearchStrings(t *testing.T) {
	// Verifica que los genéricos funcionan con cmp.Ordered no numérico.
	root := &Node[string]{numKeys: 2}
	root.keys[0] = "lima"
	root.keys[1] = "quito"
	tr := &Tree23[string]{root: root, size: 2}

	if !tr.Search("lima") {
		t.Errorf(`Search("lima") = false; se esperaba true`)
	}
	if !tr.Search("quito") {
		t.Errorf(`Search("quito") = false; se esperaba true`)
	}
	if tr.Search("bogota") {
		t.Errorf(`Search("bogota") = true; se esperaba false`)
	}
}

// --- Paso 2: pruebas de Insert e InOrder ---------------------------------

// checkInvariants verifica las dos propiedades centrales del árbol 2-3:
//  1. Orden: dentro de cada nodo keys[0] < keys[1], y cada subárbol respeta
//     los rangos definidos por las claves del padre.
//  2. Balance: todas las hojas están exactamente al mismo nivel.
//
// Devuelve la altura del subárbol (en niveles) para verificar el balance.
func checkInvariants[K int | string](t *testing.T, n *Node[K]) int {
	t.Helper()
	if n == nil {
		return 0
	}

	if n.numKeys < 1 || n.numKeys > 2 {
		t.Fatalf("numKeys inválido: %d (debe ser 1 o 2)", n.numKeys)
	}
	if n.numKeys == 2 && !(n.keys[0] < n.keys[1]) {
		t.Fatalf("claves desordenadas en nodo: %v >= %v", n.keys[0], n.keys[1])
	}

	if n.isLeaf() {
		return 1
	}

	// Un nodo interno debe tener exactamente numKeys+1 hijos no-nil.
	expectedChildren := n.numKeys + 1
	heights := make([]int, 0, expectedChildren)
	for i := 0; i < expectedChildren; i++ {
		if n.children[i] == nil {
			t.Fatalf("hijo %d nil en nodo interno con %d claves", i, n.numKeys)
		}
		heights = append(heights, checkInvariants(t, n.children[i]))
	}
	// Ningún hijo extra debe colgar más allá de expectedChildren.
	if expectedChildren < 3 && n.children[2] != nil {
		t.Fatalf("hijo extra inesperado en un 2-nodo")
	}

	// Todas las alturas de los hijos deben coincidir (árbol balanceado).
	for _, h := range heights {
		if h != heights[0] {
			t.Fatalf("alturas desiguales entre hijos: %v (árbol desbalanceado)", heights)
		}
	}
	return heights[0] + 1
}

func TestInsertSingleAndLen(t *testing.T) {
	tr := New[int]()
	tr.Insert(42)
	if !tr.Search(42) {
		t.Errorf("tras Insert(42), Search(42) = false")
	}
	if tr.Len() != 1 {
		t.Errorf("Len() = %d; se esperaba 1", tr.Len())
	}
}

func TestInsertInOrderSorted(t *testing.T) {
	tr := New[int]()
	input := []int{50, 30, 70, 20, 40, 60, 80, 10, 25, 35, 45}
	for _, k := range input {
		tr.Insert(k)
	}

	got := tr.InOrder()
	want := append([]int(nil), input...)
	sort.Ints(want)

	if len(got) != len(want) {
		t.Fatalf("InOrder() len = %d; se esperaba %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("InOrder() = %v; se esperaba %v", got, want)
		}
	}
	if tr.Len() != len(want) {
		t.Errorf("Len() = %d; se esperaba %d", tr.Len(), len(want))
	}
	checkInvariants(t, tr.root)
}

func TestInsertDuplicatesIgnored(t *testing.T) {
	tr := New[int]()
	for _, k := range []int{5, 5, 5, 3, 3, 8} {
		tr.Insert(k)
	}
	want := []int{3, 5, 8}
	got := tr.InOrder()
	if len(got) != len(want) {
		t.Fatalf("InOrder() = %v; se esperaba %v (duplicados ignorados)", got, want)
	}
	if tr.Len() != 3 {
		t.Errorf("Len() = %d; se esperaba 3", tr.Len())
	}
}

func TestInsertAscendingAndDescending(t *testing.T) {
	// Las secuencias ordenadas son el peor caso para árboles no balanceados:
	// aquí deben seguir produciendo un árbol balanceado.
	for _, name := range []string{"asc", "desc"} {
		tr := New[int]()
		const n = 200
		for i := 0; i < n; i++ {
			if name == "asc" {
				tr.Insert(i)
			} else {
				tr.Insert(n - i)
			}
		}
		if tr.Len() != n {
			t.Errorf("[%s] Len() = %d; se esperaba %d", name, tr.Len(), n)
		}
		got := tr.InOrder()
		for i := 1; i < len(got); i++ {
			if got[i-1] >= got[i] {
				t.Fatalf("[%s] InOrder no está ordenado en índice %d", name, i)
			}
		}
		checkInvariants(t, tr.root)
	}
}

func TestInsertRandomizedStress(t *testing.T) {
	tr := New[int]()
	rng := rand.New(rand.NewSource(42))
	present := make(map[int]bool)
	const ops = 2000

	for i := 0; i < ops; i++ {
		k := rng.Intn(500)
		tr.Insert(k)
		present[k] = true
	}

	// Todas las claves insertadas deben encontrarse; las no insertadas, no.
	for k := 0; k < 500; k++ {
		if tr.Search(k) != present[k] {
			t.Fatalf("Search(%d) = %v; se esperaba %v", k, tr.Search(k), present[k])
		}
	}
	if tr.Len() != len(present) {
		t.Errorf("Len() = %d; se esperaba %d", tr.Len(), len(present))
	}

	// InOrder debe estar ordenado y coincidir con el conjunto esperado.
	want := make([]int, 0, len(present))
	for k := range present {
		want = append(want, k)
	}
	sort.Ints(want)
	got := tr.InOrder()
	if len(got) != len(want) {
		t.Fatalf("InOrder() len = %d; se esperaba %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("InOrder desincronizado en índice %d: %d != %d", i, got[i], want[i])
		}
	}
	checkInvariants(t, tr.root)
}

func TestRangeQuery(t *testing.T) {
	tr := New[int]()
	for _, k := range []int{50, 30, 70, 20, 40, 60, 80, 10, 25, 35, 45} {
		tr.Insert(k)
	}

	cases := []struct {
		lo, hi int
		want   []int
	}{
		{25, 45, []int{25, 30, 35, 40, 45}}, // rango intermedio
		{0, 100, []int{10, 20, 25, 30, 35, 40, 45, 50, 60, 70, 80}}, // todo
		{33, 33, []int{}},        // extremos iguales, sin coincidencia
		{35, 35, []int{35}},      // extremos iguales, con coincidencia
		{100, 200, []int{}},      // fuera por arriba
		{-10, 5, []int{}},        // fuera por abajo
		{45, 25, []int{}},        // lo > hi
		{41, 59, []int{45, 50}},  // bordes no inclusivos de claves
	}

	for _, c := range cases {
		got := tr.RangeQuery(c.lo, c.hi)
		if len(got) != len(c.want) {
			t.Fatalf("RangeQuery(%d,%d) = %v; se esperaba %v", c.lo, c.hi, got, c.want)
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Fatalf("RangeQuery(%d,%d) = %v; se esperaba %v", c.lo, c.hi, got, c.want)
			}
		}
	}
}

func TestRangeQueryMatchesInOrderFilter(t *testing.T) {
	// Propiedad: RangeQuery(lo,hi) == InOrder filtrado por [lo,hi], para muchos
	// rangos aleatorios sobre un árbol grande.
	tr := New[int]()
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 1000; i++ {
		tr.Insert(rng.Intn(5000))
	}
	all := tr.InOrder()

	for trial := 0; trial < 100; trial++ {
		lo := rng.Intn(5000)
		hi := lo + rng.Intn(1000)

		var want []int
		for _, k := range all {
			if k >= lo && k <= hi {
				want = append(want, k)
			}
		}
		got := tr.RangeQuery(lo, hi)
		if len(got) != len(want) {
			t.Fatalf("RangeQuery(%d,%d) len = %d; se esperaba %d", lo, hi, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("RangeQuery(%d,%d) desincronizado en %d", lo, hi, i)
			}
		}
	}
}

func TestInsertStringsBalanced(t *testing.T) {
	tr := New[string]()
	ciudades := []string{"lima", "quito", "bogota", "santiago", "caracas", "montevideo", "asuncion"}
	for _, c := range ciudades {
		tr.Insert(c)
	}
	got := tr.InOrder()
	want := append([]string(nil), ciudades...)
	sort.Strings(want)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("InOrder() = %v; se esperaba %v", got, want)
		}
	}
	checkInvariants(t, tr.root)
}
