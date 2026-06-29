// Package tree23 implementa un Árbol 2-3 genérico desde cero.
//
// Un Árbol 2-3 es un árbol de búsqueda balanceado en el que cada nodo interno
// es uno de dos tipos:
//
//   - 2-nodo: contiene 1 clave y tiene 2 hijos (izquierdo y derecho).
//   - 3-nodo: contiene 2 claves y tiene 3 hijos (izquierdo, medio y derecho).
//
// Invariante de orden (para un 3-nodo con claves k0 < k1):
//
//	hijos[0] < k0 < hijos[1] < k1 < hijos[2]
//
// Todas las hojas están al mismo nivel, lo que garantiza altura O(log n) y, por
// lo tanto, búsqueda, inserción y eliminación en O(log n).
//
// Referencia: Aho, Hopcroft & Ullman (1974), The Design and Analysis of
// Computer Algorithms.
package tree23

import "cmp"

// Node es un nodo del árbol 2-3.
//
// El tipo de clave K se restringe a cmp.Ordered (enteros, flotantes y strings),
// de modo que podemos comparar claves con los operadores <, >, == directamente.
//
// Representamos tanto al 2-nodo como al 3-nodo con la misma struct usando
// arreglos de tamaño fijo y un contador numKeys:
//
//   - numKeys == 1  -> 2-nodo: usa keys[0] y children[0..1].
//   - numKeys == 2  -> 3-nodo: usa keys[0..1] y children[0..2].
//
// En una hoja todos los punteros de children son nil.
type Node[K cmp.Ordered] struct {
	keys     [2]K        // claves ordenadas: keys[0] < keys[1] (si numKeys == 2)
	children [3]*Node[K] // hijos; children[i] cuelga entre keys[i-1] y keys[i]
	numKeys  int         // 1 => 2-nodo, 2 => 3-nodo
}

// isLeaf indica si el nodo es una hoja (no tiene hijos).
//
// Como todas las hojas están al mismo nivel, basta con mirar el primer hijo:
// si children[0] es nil, el nodo no tiene hijos.
func (n *Node[K]) isLeaf() bool {
	return n.children[0] == nil
}

// Tree23 es el árbol 2-3 propiamente dicho.
//
// Mantiene un puntero a la raíz y el número de claves almacenadas (size) para
// responder Len() en O(1). Un árbol vacío tiene root == nil y size == 0.
type Tree23[K cmp.Ordered] struct {
	root *Node[K]
	size int
}

// New crea y devuelve un árbol 2-3 vacío.
func New[K cmp.Ordered]() *Tree23[K] {
	return &Tree23[K]{}
}

// Len devuelve la cantidad de claves almacenadas en el árbol.
func (t *Tree23[K]) Len() int {
	return t.size
}

// Search devuelve true si la clave existe en el árbol.
//
// Navega desde la raíz hacia abajo en O(log n): en cada nodo compara contra sus
// 1 o 2 claves para decidir si encontró la clave o por cuál hijo descender.
// Es una operación de solo lectura: no modifica el árbol.
func (t *Tree23[K]) Search(key K) bool {
	return search(t.root, key)
}

// search recorre recursivamente el subárbol con raíz n buscando key.
func search[K cmp.Ordered](n *Node[K], key K) bool {
	if n == nil {
		return false
	}

	// ¿La clave está en este nodo?
	if key == n.keys[0] {
		return true
	}
	if n.numKeys == 2 && key == n.keys[1] {
		return true
	}

	// Si es hoja y no estaba aquí, no existe.
	if n.isLeaf() {
		return false
	}

	// Decidir por cuál hijo descender según el orden.
	switch {
	case key < n.keys[0]:
		// Menor que la primera clave -> hijo izquierdo.
		return search(n.children[0], key)
	case n.numKeys == 1 || key < n.keys[1]:
		// 2-nodo: mayor que la única clave -> hijo derecho.
		// 3-nodo: entre keys[0] y keys[1] -> hijo del medio.
		return search(n.children[1], key)
	default:
		// 3-nodo: mayor que la segunda clave -> hijo derecho.
		return search(n.children[2], key)
	}
}

// split representa el resultado de dividir un nodo que se desbordó.
//
// Cuando un nodo recibe una clave de más (quedaría con 3 claves), se parte en
// dos: la clave del medio "sube" (promoted) hacia el padre, el nodo original se
// queda como mitad izquierda y se crea un nuevo nodo right como mitad derecha.
// El padre debe absorber promoted y right. Si la propagación llega hasta la
// raíz, se crea una raíz nueva y el árbol crece un nivel en altura.
type split[K cmp.Ordered] struct {
	promoted K
	right    *Node[K]
}

// Insert agrega key al árbol manteniendo el balance. Si la clave ya existe, no
// hace nada (no se permiten duplicados). Complejidad: O(log n).
func (t *Tree23[K]) Insert(key K) {
	if t.root == nil {
		// Árbol vacío: la nueva clave es la raíz (un 2-nodo).
		root := &Node[K]{numKeys: 1}
		root.keys[0] = key
		t.root = root
		t.size++
		return
	}

	s, inserted := insert(t.root, key)
	if inserted {
		t.size++
	}
	if s != nil {
		// La raíz se dividió: creamos una raíz nueva con la clave promovida.
		// Es la única forma en que un árbol 2-3 aumenta su altura.
		newRoot := &Node[K]{numKeys: 1}
		newRoot.keys[0] = s.promoted
		newRoot.children[0] = t.root
		newRoot.children[1] = s.right
		t.root = newRoot
	}
}

// insert agrega key en el subárbol con raíz n.
//
// Devuelve:
//   - *split: no-nil si n se desbordó y se dividió; el padre debe absorberlo.
//   - bool:   true si se insertó una clave nueva, false si ya existía.
func insert[K cmp.Ordered](n *Node[K], key K) (*split[K], bool) {
	if n.isLeaf() {
		if key == n.keys[0] || (n.numKeys == 2 && key == n.keys[1]) {
			return nil, false // duplicado
		}
		return addKeyToLeaf(n, key), true
	}

	// Nodo interno: elegir el hijo por el que descender (o detectar duplicado).
	var idx int
	switch {
	case key == n.keys[0]:
		return nil, false
	case key < n.keys[0]:
		idx = 0
	case n.numKeys == 1:
		idx = 1
	case key == n.keys[1]:
		return nil, false
	case key < n.keys[1]:
		idx = 1
	default:
		idx = 2
	}

	s, inserted := insert(n.children[idx], key)
	if s == nil {
		return nil, inserted // el hijo absorbió la clave sin dividirse
	}
	// El hijo idx se dividió: absorbemos la clave promovida y el nuevo hijo.
	return addChildToInternal(n, idx, s), inserted
}

// addKeyToLeaf inserta key en una hoja, manteniendo las claves ordenadas.
//
// Si la hoja era un 2-nodo, pasa a ser un 3-nodo y no hay desbordamiento.
// Si era un 3-nodo, quedaría con 3 claves: se divide y se devuelve el split.
func addKeyToLeaf[K cmp.Ordered](n *Node[K], key K) *split[K] {
	if n.numKeys == 1 {
		// 2-nodo -> 3-nodo: ubicar key respecto a la única clave existente.
		if key < n.keys[0] {
			n.keys[1] = n.keys[0]
			n.keys[0] = key
		} else {
			n.keys[1] = key
		}
		n.numKeys = 2
		return nil
	}

	// 3-nodo + 1 clave = desbordamiento. Ordenamos las tres claves; como
	// keys[0] < keys[1] ya se cumple, solo ubicamos la nueva (key).
	a, b := n.keys[0], n.keys[1]
	var lo, mid, hi K
	switch {
	case key < a:
		lo, mid, hi = key, a, b
	case key < b:
		lo, mid, hi = a, key, b
	default:
		lo, mid, hi = a, b, key
	}

	// n se queda como mitad izquierda (solo lo); hi va a un nodo nuevo; mid sube.
	n.keys[0] = lo
	var zero K
	n.keys[1] = zero
	n.numKeys = 1

	right := &Node[K]{numKeys: 1}
	right.keys[0] = hi
	return &split[K]{promoted: mid, right: right}
}

// addChildToInternal absorbe en el nodo interno n el resultado de la división
// de su hijo en la posición idx: inserta s.promoted entre las claves y s.right
// entre los hijos.
//
// Si n era un 2-nodo, pasa a 3-nodo sin desbordarse. Si era un 3-nodo, quedaría
// con 3 claves y 4 hijos: se vuelve a dividir y el split sube otro nivel.
func addChildToInternal[K cmp.Ordered](n *Node[K], idx int, s *split[K]) *split[K] {
	if n.numKeys == 1 {
		// 2-nodo -> 3-nodo. idx es 0 o 1 (el hijo dividido).
		if idx == 0 {
			n.keys[1] = n.keys[0]
			n.keys[0] = s.promoted
			n.children[2] = n.children[1]
			n.children[1] = s.right
		} else { // idx == 1
			n.keys[1] = s.promoted
			n.children[2] = s.right
		}
		n.numKeys = 2
		return nil
	}

	// 3-nodo desbordado: construimos las 3 claves y 4 hijos resultantes en
	// orden, insertando s.promoted y s.right según la posición idx.
	var keys [3]K
	var ch [4]*Node[K]
	switch idx {
	case 0:
		keys = [3]K{s.promoted, n.keys[0], n.keys[1]}
		ch = [4]*Node[K]{n.children[0], s.right, n.children[1], n.children[2]}
	case 1:
		keys = [3]K{n.keys[0], s.promoted, n.keys[1]}
		ch = [4]*Node[K]{n.children[0], n.children[1], s.right, n.children[2]}
	default: // idx == 2
		keys = [3]K{n.keys[0], n.keys[1], s.promoted}
		ch = [4]*Node[K]{n.children[0], n.children[1], n.children[2], s.right}
	}

	// La clave del medio (keys[1]) sube. n se queda como mitad izquierda
	// (keys[0] + dos hijos) y creamos right como mitad derecha.
	var zero K
	n.keys[0] = keys[0]
	n.keys[1] = zero
	n.numKeys = 1
	n.children[0] = ch[0]
	n.children[1] = ch[1]
	n.children[2] = nil

	right := &Node[K]{numKeys: 1}
	right.keys[0] = keys[2]
	right.children[0] = ch[2]
	right.children[1] = ch[3]

	return &split[K]{promoted: keys[1], right: right}
}

// InOrder devuelve todas las claves del árbol en orden ascendente.
//
// Recorre el árbol intercalando hijos y claves, lo que produce una secuencia
// ordenada en O(n). Es la base de los listados ordenados y las consultas por
// rango del Entregable 3.
func (t *Tree23[K]) InOrder() []K {
	result := make([]K, 0, t.size)
	inOrder(t.root, &result)
	return result
}

func inOrder[K cmp.Ordered](n *Node[K], out *[]K) {
	if n == nil {
		return
	}
	if n.isLeaf() {
		*out = append(*out, n.keys[0])
		if n.numKeys == 2 {
			*out = append(*out, n.keys[1])
		}
		return
	}

	// 2-nodo: hijo0, clave0, hijo1.
	// 3-nodo: hijo0, clave0, hijo1, clave1, hijo2.
	inOrder(n.children[0], out)
	*out = append(*out, n.keys[0])
	inOrder(n.children[1], out)
	if n.numKeys == 2 {
		*out = append(*out, n.keys[1])
		inOrder(n.children[2], out)
	}
}

// RangeQuery devuelve, en orden ascendente, todas las claves k del árbol con
// lo <= k <= hi (ambos extremos incluidos). Si lo > hi, devuelve vacío.
//
// A diferencia de InOrder, poda las ramas que quedan completamente fuera del
// rango: no desciende por un hijo si todas sus claves serían menores que lo o
// mayores que hi. Costo O(log n + m), con m = cantidad de claves devueltas.
//
// Es la consulta central del Entregable 3 (p. ej. "ciudades alfabéticamente
// entre A y B") y donde el árbol ordenado luce frente a un hash o un slice.
func (t *Tree23[K]) RangeQuery(lo, hi K) []K {
	result := make([]K, 0)
	if lo > hi {
		return result
	}
	rangeQuery(t.root, lo, hi, &result)
	return result
}

func rangeQuery[K cmp.Ordered](n *Node[K], lo, hi K, out *[]K) {
	if n == nil {
		return
	}

	if n.isLeaf() {
		if n.keys[0] >= lo && n.keys[0] <= hi {
			*out = append(*out, n.keys[0])
		}
		if n.numKeys == 2 && n.keys[1] >= lo && n.keys[1] <= hi {
			*out = append(*out, n.keys[1])
		}
		return
	}

	// Hijo izquierdo: solo si pueden existir claves >= lo (es decir, lo < keys[0]).
	if lo < n.keys[0] {
		rangeQuery(n.children[0], lo, hi, out)
	}
	// Primera clave del nodo.
	if n.keys[0] >= lo && n.keys[0] <= hi {
		*out = append(*out, n.keys[0])
	}

	if n.numKeys == 1 {
		// 2-nodo: hijo derecho solo si pueden existir claves <= hi (hi > keys[0]).
		if hi > n.keys[0] {
			rangeQuery(n.children[1], lo, hi, out)
		}
		return
	}

	// 3-nodo: hijo del medio si el rango se cruza con (keys[0], keys[1]).
	if hi > n.keys[0] && lo < n.keys[1] {
		rangeQuery(n.children[1], lo, hi, out)
	}
	// Segunda clave del nodo.
	if n.keys[1] >= lo && n.keys[1] <= hi {
		*out = append(*out, n.keys[1])
	}
	// Hijo derecho si pueden existir claves <= hi (hi > keys[1]).
	if hi > n.keys[1] {
		rangeQuery(n.children[2], lo, hi, out)
	}
}
