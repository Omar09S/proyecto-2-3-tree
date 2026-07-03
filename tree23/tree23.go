package tree23

import (
	"cmp"
	"fmt"
)

type Node[K cmp.Ordered] struct {
	keys     [3]K
	children [4]*Node[K]
	numKeys  int
}

func (n *Node[K]) isLeaf() bool {
	return n.children[0] == nil
}

type Tree23[K cmp.Ordered] struct {
	root *Node[K]
	size int
}

func New[K cmp.Ordered]() *Tree23[K] {
	return &Tree23[K]{}
}

func newNode[K cmp.Ordered](key K) *Node[K] {
	n := &Node[K]{numKeys: 1}
	n.keys[0] = key
	return n
}

func (t *Tree23[K]) Len() int {
	return t.size
}

func (t *Tree23[K]) Search(key K) bool {
	return search(t.root, key)
}

func search[K cmp.Ordered](node *Node[K], key K) bool {
	if node == nil {
		return false
	}
	if key == node.keys[0] {
		return true
	}
	if node.numKeys == 2 && key == node.keys[1] {
		return true
	}
	if node.isLeaf() {
		return false
	}
	switch {
	case key < node.keys[0]:
		return search(node.children[0], key)
	case node.numKeys == 1 || key < node.keys[1]:
		return search(node.children[1], key)
	default:
		return search(node.children[2], key)
	}
}

func (t *Tree23[K]) Insert(key K) {
	if t.root == nil {
		t.root = newNode(key)
		t.size++
		return
	}

	var inserted bool
	t.root = t.insert(t.root, key, &inserted)

	if t.root.numKeys == 3 {
		newRoot := &Node[K]{}
		newRoot.children[0] = t.root
		t.root = t.split(newRoot)
	}
	if inserted {
		t.size++
	}
}

func (t *Tree23[K]) insert(node *Node[K], key K, inserted *bool) *Node[K] {
	if node.isLeaf() {
		return t.insertIntoLeaf(node, key, inserted)
	}

	if key == node.keys[0] || (node.numKeys == 2 && key == node.keys[1]) {
		return node
	}

	if key < node.keys[0] {
		node.children[0] = t.insert(node.children[0], key, inserted)
	} else if node.numKeys == 1 || key < node.keys[1] {
		node.children[1] = t.insert(node.children[1], key, inserted)
	} else {
		node.children[2] = t.insert(node.children[2], key, inserted)
	}

	if node.children[0] != nil && node.children[0].numKeys == 3 {
		node = t.split(node)
	} else if node.children[1] != nil && node.children[1].numKeys == 3 {
		node = t.split(node)
	} else if node.children[2] != nil && node.children[2].numKeys == 3 {
		node = t.split(node)
	}
	return node
}

func (t *Tree23[K]) insertIntoLeaf(node *Node[K], key K, inserted *bool) *Node[K] {
	if key == node.keys[0] || (node.numKeys == 2 && key == node.keys[1]) {
		return node
	}
	*inserted = true

	if node.numKeys == 1 {
		if key < node.keys[0] {
			node.keys[1] = node.keys[0]
			node.keys[0] = key
		} else {
			node.keys[1] = key
		}
		node.numKeys = 2
		return node
	}

	if key < node.keys[0] {
		node.keys = [3]K{key, node.keys[0], node.keys[1]}
	} else if key < node.keys[1] {
		node.keys = [3]K{node.keys[0], key, node.keys[1]}
	} else {
		node.keys = [3]K{node.keys[0], node.keys[1], key}
	}
	node.numKeys = 3
	return node
}

func (t *Tree23[K]) split(node *Node[K]) *Node[K] {
	var overflow int
	for i := 0; i < 4; i++ {
		if node.children[i] != nil && node.children[i].numKeys == 3 {
			overflow = i
			break
		}
	}

	child := node.children[overflow]
	leftKey := child.keys[0]
	midKey := child.keys[1]
	rightKey := child.keys[2]

	left := newNode(leftKey)
	right := newNode(rightKey)
	left.children[0] = child.children[0]
	left.children[1] = child.children[1]
	right.children[0] = child.children[2]
	right.children[1] = child.children[3]

	if node.numKeys == 0 {
		node.keys[0] = midKey
		node.numKeys = 1
		node.children[0] = left
		node.children[1] = right
		return node
	}

	if midKey < node.keys[0] {
		node.keys[2] = node.keys[1]
		node.keys[1] = node.keys[0]
		node.keys[0] = midKey
		node.children[3] = node.children[2]
		node.children[2] = node.children[1]
		node.children[1] = right
		node.children[0] = left
	} else if node.numKeys == 1 || midKey < node.keys[1] {
		node.keys[2] = node.keys[1]
		node.keys[1] = midKey
		node.children[3] = node.children[2]
		node.children[2] = right
		node.children[overflow] = left
	} else {
		node.keys[2] = midKey
		node.children[3] = right
		node.children[overflow] = left
	}
	node.numKeys++
	return node
}

func (t *Tree23[K]) Delete(key K) {
	if t.root == nil {
		return
	}
	var removed bool
	t.root = t.delete(t.root, key, &removed)

	if t.root != nil && t.root.numKeys == 0 {
		t.root = t.root.children[0]
	}
	if removed {
		t.size--
	}
}

func (t *Tree23[K]) delete(node *Node[K], key K, removed *bool) *Node[K] {
	if node == nil {
		return nil
	}

	if node.isLeaf() {
		if t.removeFromLeaf(node, key) {
			*removed = true
		}
		return node
	}

	if key == node.keys[0] {
		*removed = true
		pred := maxKey(node.children[0])
		node.keys[0] = pred
		node.children[0] = t.delete(node.children[0], pred, new(bool))
		if node.children[0] != nil && node.children[0].numKeys == 0 {
			node = t.fix(node, 0)
		}
		return node
	}
	if node.numKeys == 2 && key == node.keys[1] {
		*removed = true
		pred := maxKey(node.children[1])
		node.keys[1] = pred
		node.children[1] = t.delete(node.children[1], pred, new(bool))
		if node.children[1] != nil && node.children[1].numKeys == 0 {
			node = t.fix(node, 1)
		}
		return node
	}

	var idx int
	if key < node.keys[0] {
		idx = 0
	} else if node.numKeys == 1 || key < node.keys[1] {
		idx = 1
	} else {
		idx = 2
	}
	node.children[idx] = t.delete(node.children[idx], key, removed)
	if node.children[idx] != nil && node.children[idx].numKeys == 0 {
		node = t.fix(node, idx)
	}
	return node
}

func (t *Tree23[K]) removeFromLeaf(node *Node[K], key K) bool {
	var zero K
	switch {
	case key == node.keys[0] && node.numKeys == 2:
		node.keys[0] = node.keys[1]
		node.keys[1] = zero
		node.numKeys--
		return true
	case node.numKeys == 2 && key == node.keys[1]:
		node.keys[1] = zero
		node.numKeys--
		return true
	case key == node.keys[0] && node.numKeys == 1:
		node.keys[0] = zero
		node.numKeys--
		return true
	}
	return false
}

func (t *Tree23[K]) fix(father *Node[K], idx int) *Node[K] {
	if idx < father.numKeys && father.children[idx+1] != nil && father.children[idx+1].numKeys == 2 {
		t.redistribute(father, idx)
	} else if idx > 0 && father.children[idx-1] != nil && father.children[idx-1].numKeys == 2 {
		t.redistribute(father, idx)
	} else {
		father = t.merge(father, idx)
	}
	return father
}

func (t *Tree23[K]) redistribute(father *Node[K], idx int) {
	child := father.children[idx]
	var zero K

	if idx < father.numKeys && father.children[idx+1] != nil && father.children[idx+1].numKeys == 2 {
		right := father.children[idx+1]
		child.keys[0] = father.keys[idx]
		child.numKeys = 1
		child.children[1] = right.children[0]

		father.keys[idx] = right.keys[0]

		right.keys[0] = right.keys[1]
		right.keys[1] = zero
		right.children[0] = right.children[1]
		right.children[1] = right.children[2]
		right.children[2] = nil
		right.numKeys--
	} else {
		left := father.children[idx-1]
		child.keys[0] = father.keys[idx-1]
		child.numKeys = 1
		child.children[1] = child.children[0]
		child.children[0] = left.children[left.numKeys]

		father.keys[idx-1] = left.keys[left.numKeys-1]

		left.keys[left.numKeys-1] = zero
		left.children[left.numKeys] = nil
		left.numKeys--
	}
}

func (t *Tree23[K]) merge(father *Node[K], idx int) *Node[K] {
	var zero K

	if idx < father.numKeys {
		child := father.children[idx]
		right := father.children[idx+1]

		child.keys[0] = father.keys[idx]
		child.keys[1] = right.keys[0]
		child.numKeys = 2
		child.children[1] = right.children[0]
		child.children[2] = right.children[1]

		for i := idx; i < father.numKeys-1; i++ {
			father.keys[i] = father.keys[i+1]
			father.children[i+1] = father.children[i+2]
		}
	} else {
		left := father.children[idx-1]
		child := father.children[idx]

		left.keys[1] = father.keys[idx-1]
		left.numKeys = 2
		left.children[2] = child.children[0]

		for i := idx - 1; i < father.numKeys-1; i++ {
			father.keys[i] = father.keys[i+1]
			father.children[i+1] = father.children[i+2]
		}
	}

	father.keys[father.numKeys-1] = zero
	father.children[father.numKeys] = nil
	father.numKeys--
	return father
}

func maxKey[K cmp.Ordered](node *Node[K]) K {
	for !node.isLeaf() {
		node = node.children[node.numKeys]
	}
	return node.keys[node.numKeys-1]
}

func (t *Tree23[K]) InOrder() []K {
	result := make([]K, 0, t.size)
	inOrder(t.root, &result)
	return result
}

func inOrder[K cmp.Ordered](node *Node[K], out *[]K) {
	if node == nil {
		return
	}
	if node.isLeaf() {
		*out = append(*out, node.keys[0])
		if node.numKeys == 2 {
			*out = append(*out, node.keys[1])
		}
		return
	}
	inOrder(node.children[0], out)
	*out = append(*out, node.keys[0])
	inOrder(node.children[1], out)
	if node.numKeys == 2 {
		*out = append(*out, node.keys[1])
		inOrder(node.children[2], out)
	}
}

func (t *Tree23[K]) RangeQuery(lo, hi K) []K {
	result := make([]K, 0)
	if lo > hi {
		return result
	}
	rangeQuery(t.root, lo, hi, &result)
	return result
}

func rangeQuery[K cmp.Ordered](node *Node[K], lo, hi K, out *[]K) {
	if node == nil {
		return
	}
	if node.isLeaf() {
		if node.keys[0] >= lo && node.keys[0] <= hi {
			*out = append(*out, node.keys[0])
		}
		if node.numKeys == 2 && node.keys[1] >= lo && node.keys[1] <= hi {
			*out = append(*out, node.keys[1])
		}
		return
	}

	if lo < node.keys[0] {
		rangeQuery(node.children[0], lo, hi, out)
	}
	if node.keys[0] >= lo && node.keys[0] <= hi {
		*out = append(*out, node.keys[0])
	}

	if node.numKeys == 1 {
		if hi > node.keys[0] {
			rangeQuery(node.children[1], lo, hi, out)
		}
		return
	}

	if hi > node.keys[0] && lo < node.keys[1] {
		rangeQuery(node.children[1], lo, hi, out)
	}
	if node.keys[1] >= lo && node.keys[1] <= hi {
		*out = append(*out, node.keys[1])
	}
	if hi > node.keys[1] {
		rangeQuery(node.children[2], lo, hi, out)
	}
}

type NodeSnapshot[K cmp.Ordered] struct {
	Keys     []K                `json:"keys"`
	NumKeys  int                `json:"numKeys"`
	IsLeaf   bool               `json:"isLeaf"`
	Children []*NodeSnapshot[K] `json:"children"`
	State    string             `json:"state"`
}

type Step[K cmp.Ordered] struct {
	Phase string           `json:"phase"`
	Msg   string           `json:"msg"`
	Case  string           `json:"case,omitempty"`
	Tree  *NodeSnapshot[K] `json:"tree"`
}

func (t *Tree23[K]) Snapshot() *NodeSnapshot[K] {
	return snapshotNode(t.root, nil)
}

func snapshotNode[K cmp.Ordered](n *Node[K], states map[*Node[K]]string) *NodeSnapshot[K] {
	if n == nil {
		return nil
	}
	state := ""
	if states != nil {
		state = states[n]
	}
	snap := &NodeSnapshot[K]{
		NumKeys:  n.numKeys,
		IsLeaf:   n.isLeaf(),
		State:    state,
		Children: make([]*NodeSnapshot[K], 3),
	}
	for i := 0; i < n.numKeys && i < 2; i++ {
		snap.Keys = append(snap.Keys, n.keys[i])
	}
	for i := 0; i < 3; i++ {
		snap.Children[i] = snapshotNode(n.children[i], states)
	}
	return snap
}

func findNodeWithKey[K cmp.Ordered](n *Node[K], key K) *Node[K] {
	if n == nil {
		return nil
	}
	if key == n.keys[0] || (n.numKeys == 2 && key == n.keys[1]) {
		return n
	}
	if n.isLeaf() {
		return nil
	}
	switch {
	case key < n.keys[0]:
		return findNodeWithKey(n.children[0], key)
	case n.numKeys == 1 || key < n.keys[1]:
		return findNodeWithKey(n.children[1], key)
	default:
		return findNodeWithKey(n.children[2], key)
	}
}

func collectPathSlice[K cmp.Ordered](n *Node[K], key K, path *[]*Node[K]) bool {
	if n == nil {
		return false
	}
	*path = append(*path, n)
	if key == n.keys[0] || (n.numKeys == 2 && key == n.keys[1]) {
		return true
	}
	if n.isLeaf() {
		return false
	}
	switch {
	case key < n.keys[0]:
		return collectPathSlice(n.children[0], key, path)
	case n.numKeys == 1 || key < n.keys[1]:
		return collectPathSlice(n.children[1], key, path)
	default:
		return collectPathSlice(n.children[2], key, path)
	}
}

func treeHeightOf[K cmp.Ordered](n *Node[K]) int {
	h := 0
	for n != nil {
		h++
		n = n.children[0]
	}
	return h
}

func (t *Tree23[K]) SearchSteps(key K) (bool, []Step[K]) {
	var steps []Step[K]
	var path []*Node[K]
	found := collectPathSlice(t.root, key, &path)

	for i, node := range path {
		states := make(map[*Node[K]]string)
		for j := 0; j < i; j++ {
			states[path[j]] = "visited"
		}
		isLast := i == len(path)-1
		if isLast && found {
			states[node] = "found"
		} else {
			states[node] = "active"
		}

		var phase, msg string
		switch {
		case i == 0:
			phase, msg = "descend", "Iniciando búsqueda desde la raíz"
		case isLast && found:
			phase, msg = "found", "¡Clave encontrada!"
		case isLast:
			phase, msg = "not_found", "Hoja alcanzada — la clave no existe en el árbol"
		default:
			phase, msg = "descend", "Descendiendo al siguiente nodo"
		}
		steps = append(steps, Step[K]{Phase: phase, Msg: msg, Tree: snapshotNode(t.root, states)})
	}
	return found, steps
}

func (t *Tree23[K]) InsertSteps(key K) []Step[K] {
	var steps []Step[K]

	if t.root == nil {
		t.Insert(key)
		states := map[*Node[K]]string{}
		if fn := findNodeWithKey(t.root, key); fn != nil {
			states[fn] = "new"
		}
		steps = append(steps, Step[K]{Phase: "insert", Msg: "Árbol vacío — nueva clave se convierte en raíz", Tree: snapshotNode(t.root, states)})
		steps = append(steps, Step[K]{Phase: "complete", Msg: fmt.Sprintf("Inserción completada. %d clave en el árbol.", t.Len()), Tree: t.Snapshot()})
		return steps
	}

	var path []*Node[K]
	alreadyExists := collectPathSlice(t.root, key, &path)

	for i, node := range path {
		states := make(map[*Node[K]]string)
		for j := 0; j < i; j++ {
			states[path[j]] = "visited"
		}
		states[node] = "active"
		isLast := i == len(path)-1
		var msg string
		switch {
		case i == 0:
			msg = "Iniciando inserción desde la raíz"
		case isLast && node.numKeys == 2:
			msg = "Hoja con 2 claves — se producirá una división al insertar"
		case isLast:
			msg = "Hoja encontrada — hay espacio para insertar directamente"
		default:
			msg = "Descendiendo hacia la hoja correcta"
		}
		steps = append(steps, Step[K]{Phase: "descend", Msg: msg, Tree: snapshotNode(t.root, states)})
	}

	if alreadyExists {
		steps = append(steps, Step[K]{Phase: "duplicate", Msg: "La clave ya existe — inserción ignorada", Case: "Caso: clave duplicada (ignorada)", Tree: t.Snapshot()})
		return steps
	}

	leafWasFull := len(path) > 0 && path[len(path)-1].numKeys == 2
	heightBefore := treeHeightOf(t.root)

	t.Insert(key)

	heightGrew := treeHeightOf(t.root) > heightBefore

	states := map[*Node[K]]string{}
	if fn := findNodeWithKey(t.root, key); fn != nil {
		states[fn] = "new"
	}

	var insertCase string
	if !leafWasFull {
		insertCase = "Caso 1 — inserción directa (hoja con espacio)"
		steps = append(steps, Step[K]{Phase: "insert", Msg: "Clave insertada directamente en la hoja (sin división)", Tree: snapshotNode(t.root, states)})
	} else if heightGrew {
		insertCase = "Caso 3 — división propagada hasta la raíz (árbol creció)"
		steps = append(steps, Step[K]{Phase: "split", Msg: "Nodo hoja lleno — se dividió y la clave media ascendió", Tree: snapshotNode(t.root, states)})
		steps = append(steps, Step[K]{Phase: "grow", Msg: "El split llegó a la raíz — el árbol creció un nivel de altura", Tree: t.Snapshot()})
	} else {
		insertCase = "Caso 2 — división de hoja (clave media sube al padre)"
		steps = append(steps, Step[K]{Phase: "split", Msg: "Nodo hoja lleno — se dividió y la clave media ascendió al padre", Tree: snapshotNode(t.root, states)})
	}

	steps = append(steps, Step[K]{Phase: "complete", Msg: fmt.Sprintf("Inserción completada. %d claves en el árbol.", t.Len()), Case: insertCase, Tree: t.Snapshot()})
	return steps
}

func (t *Tree23[K]) DeleteSteps(key K) []Step[K] {
	var steps []Step[K]
	var path []*Node[K]
	found := collectPathSlice(t.root, key, &path)

	for i, node := range path {
		states := make(map[*Node[K]]string)
		for j := 0; j < i; j++ {
			states[path[j]] = "visited"
		}
		isLast := i == len(path)-1
		if isLast && found {
			states[node] = "found"
		} else {
			states[node] = "active"
		}
		var phase, msg string
		switch {
		case i == 0:
			phase, msg = "descend", "Iniciando eliminación desde la raíz"
		case isLast && found:
			phase, msg = "found", "Clave encontrada — procediendo a eliminar"
		case isLast:
			phase, msg = "not_found", "Clave no encontrada en el árbol"
		default:
			phase, msg = "descend", "Descendiendo hacia la clave"
		}
		steps = append(steps, Step[K]{Phase: phase, Msg: msg, Tree: snapshotNode(t.root, states)})
	}

	if !found {
		steps = append(steps, Step[K]{Phase: "not_found", Msg: "La clave no existe — nada que eliminar", Tree: t.Snapshot()})
		return steps
	}

	foundNode := path[len(path)-1]
	isInternal := !foundNode.isLeaf()
	leafWas3Node := !isInternal && foundNode.numKeys == 2

	if isInternal {
		states := map[*Node[K]]string{foundNode: "active"}
		steps = append(steps, Step[K]{Phase: "successor", Msg: "Clave en nodo interno — se reemplaza por el predecesor in-order antes de eliminar", Tree: snapshotNode(t.root, states)})
	}

	heightBefore := treeHeightOf(t.root)
	t.Delete(key)
	heightDecreased := treeHeightOf(t.root) < heightBefore

	var deleteCase string
	switch {
	case isInternal:
		deleteCase = "Caso 4 — nodo interno: reemplazo por predecesor in-order"
	case leafWas3Node:
		deleteCase = "Caso 1 — eliminación directa (3-nodo → 2-nodo)"
	case heightDecreased:
		deleteCase = "Caso 3 — fusión en cadena hasta la raíz (árbol decreció)"
	default:
		deleteCase = "Caso 2 — rotación o fusión local (altura constante)"
	}

	if heightDecreased {
		steps = append(steps, Step[K]{Phase: "merge", Msg: "Fusión llegó a la raíz — el árbol decreció un nivel de altura", Tree: t.Snapshot()})
	}

	steps = append(steps, Step[K]{Phase: "complete", Msg: fmt.Sprintf("Eliminación completada. %d claves en el árbol.", t.Len()), Case: deleteCase, Tree: t.Snapshot()})
	return steps
}
