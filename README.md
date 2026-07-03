# Árbol 2-3 — Implementación en Go

**Universidad:** ESAN  
**Curso:** Algoritmos y Estructura de Datos  
**Ciclo:** 2026-1  
**Profesor:** Calderón Niquin, Marks  

**Integrantes:**
| Apellidos y Nombres | Código |
|---|---|
| Sanchez Quispe, Omar | 25200349 |
| Mollinedo Camargo, Reginaldo | 25200487 |

**Video:** https://youtu.be/Kl46933322I

---

## ¿Qué es un Árbol 2-3?

Un **Árbol 2-3** es un árbol de búsqueda balanceado en el que cada nodo interno puede tener dos formas:

- **2-nodo:** contiene **1 clave** y tiene **2 hijos**.
- **3-nodo:** contiene **2 claves** y tiene **3 hijos**.

La propiedad más importante: **todas las hojas están siempre al mismo nivel**. Esto garantiza que el árbol permanezca perfectamente balanceado sin importar el orden de inserción, lo que a su vez garantiza que Search, Insert y RangeQuery sean siempre O(log n) en el peor caso.

```
Ejemplo con 5 claves — árbol de altura 2:

         [ 10 | 20 ]          ← raíz: 3-nodo
        /      |      \
      [5]    [15]    [25]      ← hojas: 2-nodos (mismo nivel)
```

> Referencia: Aho, Hopcroft & Ullman (1974), *The Design and Analysis of Computer Algorithms*.

---

## Estructura del repositorio

```
proyecto-2-3-tree/
├── tree23/                  # Entregable 2 — paquete central
│   ├── tree23.go            # Implementación del árbol (genéricos)
│   ├── tree23_test.go       # 21 pruebas unitarias
│   └── tree23_bench_test.go # Benchmarks: Insert, Search, RangeQuery
├── cmd/
│   ├── dbapp/main.go        # Entregable 3 — app con SQLite
│   └── api/main.go          # Entregable 4 — API REST
├── web/
│   └── index.html           # Entregable 4 — frontend Vue 3
├── data/
│   └── ciudades.csv         # Dataset: 50 ciudades latinoamericanas
└── go.mod
```

---

## Stack tecnológico

| Componente | Tecnología |
|---|---|
| Estructura del árbol | Go 1.26.1 — genéricos (`cmp.Ordered`) |
| Persistencia | SQLite (via `modernc.org/sqlite`, sin CGO) |
| API REST | Go — `net/http` estándar |
| Frontend | Vue 3 (CDN) + SVG |

---

## Cómo ejecutar

### Entregable 2 — Tests y benchmarks

```bash
# Pruebas unitarias
go test ./tree23/ -v

# Benchmarks (Search, Insert, RangeQuery en 1k / 10k / 100k claves)
go test ./tree23/ -bench=. -benchmem -run='^$'
```

### Entregable 3 — App con SQLite

```bash
go run ./cmd/dbapp
```

Carga `data/ciudades.csv` en `data/ciudades.db` y ejecuta búsqueda exacta, rango alfabético y listado in-order.

### Entregable 4 — Simulación (API + Vue)

```bash
go run ./cmd/api
```

Abrir **http://localhost:8080** en el navegador.

---

## Análisis de complejidad Big-O

### Altura del árbol

Un Árbol 2-3 con *n* claves tiene altura entre log₃(n+1) y log₂(n):

```
Altura mínima: ⌈log₃(n+1)⌉   (todos los nodos son 3-nodos)
Altura máxima: ⌊log₂(n)⌋     (todos los nodos son 2-nodos)

Por ejemplo, con n = 1 000 000:
  Mínima: ≈ 13 niveles
  Máxima: ≈ 20 niveles
```

Esto acota la altura a **Θ(log n)**, lo que garantiza las complejidades siguientes.

---

### Search — O(log n)

```
Entrada: clave k, árbol de n claves.
Por cada nivel: comparar k contra 1 o 2 claves del nodo → O(1).
Niveles recorridos: ≤ altura = O(log n).
Total: O(log n)   — peor, promedio y mejor caso.

Verificación empírica (benchmarks, i5-12450HX):
  n = 1 000   →   61 ns/op
  n = 10 000  →  212 ns/op   (×10 datos → ×3.5 tiempo)
  n = 100 000 →  382 ns/op   (×10 datos → ×1.8 tiempo)

Crecimiento logarítmico: 0 allocations/op (solo navegación de punteros).
```

---

### Insert — O(log n)

La inserción tiene dos fases:

**Fase 1 — descenso:** igual que Search, O(log n).

**Fase 2 — splits (divisiones):** cuando una hoja queda con 3 claves, se parte: la clave del medio **sube** al padre. Si el padre también se llena, se vuelve a partir. En el peor caso la propagación llega a la raíz, pero nunca supera la altura del árbol.

```
Splits posibles por inserción: ≤ altura = O(log n).
Cada split: O(1) (reordenar 3 claves y 4 punteros).
Total: O(log n)

El árbol crece de altura SOLO cuando la raíz se parte → el
incremento de altura ocurre en O(1) amortizado sobre n inserciones.

Construir el árbol desde n claves: O(n log n).

Verificación empírica (construcción de n claves):
  n = 1 000   → 0.26 ms  (260 ns/clave promedio)
  n = 10 000  → 5.4 ms   (540 ns/clave promedio)
  n = 100 000 → 80 ms    (800 ns/clave promedio)

Cada división (split) crea dos nodos nuevos, por lo que Insert sí asigna
memoria (a diferencia de Search/RangeQuery, que son 0 alloc). El crecimiento
por clave sigue siendo logarítmico.
```

---

### InOrder — O(n)

```
Recorre exactamente cada clave una vez, sin saltos ni backtracking.
Total: Θ(n).
```

---

### RangeQuery(lo, hi) — O(log n + m)

El RangeQuery **poda** las ramas que quedan completamente fuera del rango:
no desciende por un hijo si todas sus claves serían < lo o > hi.

```
Costo de la poda (localizar lo): O(log n).
Costo de recoger los m resultados:  O(m).
Total: O(log n + m).

Si m = 0 (rango vacío): O(log n), más eficiente que escanear todo.
Si m = n (rango total): O(n), equivalente a InOrder.

Verificación empírica (rango acotado, m pequeño):
  n = 1 000   →  147 ns/op
  n = 10 000  →  282 ns/op
  n = 100 000 →  487 ns/op

0 allocations/op en estas mediciones.
```

---

### Tabla resumen

| Operación | Peor caso | Promedio | Espacio extra |
|---|---|---|---|
| Search | O(log n) | O(log n) | O(1) |
| Insert | O(log n) | O(log n) | O(1) por split |
| InOrder | O(n) | O(n) | O(n) salida |
| RangeQuery | O(log n + m) | O(log n + m) | O(m) salida |
| Espacio total | — | — | O(n) |

---

### Comparación con otras estructuras

| Estructura | Search | Insert | RangeQuery | Balance garantizado |
|---|---|---|---|---|
| **Árbol 2-3** | O(log n) | O(log n) | O(log n + m) | ✅ Siempre |
| BST sin balance | O(n) peor | O(n) peor | O(n) | ❌ No |
| AVL / Rojo-Negro | O(log n) | O(log n) | O(log n + m) | ✅ Siempre |
| Hash Map | O(1) prom. | O(1) prom. | ❌ No soporta | — |
| Arreglo ordenado | O(log n) | O(n) | O(log n + m) | — |

**Ventaja clave frente al hash map:** el hash map no puede responder consultas por rango ordenadas (`RangeQuery`) — tendría que escanear todos los elementos (O(n)). El Árbol 2-3 lo hace en O(log n + m).

**Ventaja frente al BST sin balance:** en el peor caso (inserción ordenada), un BST degenera en una lista enlazada de altura n. El Árbol 2-3 mantiene la altura logarítmica **siempre**, por construcción (splits).

---

## Decisiones de implementación

### Genéricos con `cmp.Ordered`

```go
type Node[K cmp.Ordered] struct { ... }
type Tree23[K cmp.Ordered] struct { ... }
```

`cmp.Ordered` (Go 1.21+) abarca `int`, `float64`, `string` y todos sus variantes. Una sola implementación sirve para el dataset de ciudades (`string`) y para los benchmarks numéricos (`int`).

### Una sola struct para 2-nodo y 3-nodo

```go
type Node[K cmp.Ordered] struct {
    keys     [3]K
    children [4]*Node[K]
    numKeys  int          // 1 = 2-nodo, 2 = 3-nodo (3 = desbordamiento transitorio)
}
```

Los arreglos son de tamaño 3 (claves) y 4 (hijos) para poder sostener un **desbordamiento transitorio** durante la inserción: la clave se agrega primero a la hoja (que puede quedar con 3 claves) y el nodo se **divide en el camino de regreso**, promoviendo la clave del medio al padre. En un árbol ya asentado `numKeys` es siempre 1 o 2. El campo `numKeys` distingue los tipos sin necesidad de una interfaz o union.

### El árbol crece desde la raíz

A diferencia de un BST que crece desde las hojas, en el Árbol 2-3 la altura aumenta **solo** cuando la raíz se divide. Este es el mecanismo que garantiza que todas las hojas queden al mismo nivel: las hojas nunca suben, es la raíz la que crea un nuevo nivel encima.

---

## Pruebas unitarias

```
go test ./tree23/ -v

Búsqueda (5)
=== TestSearchEmptyTree               PASS
=== TestSearchSingleKey               PASS
=== TestSearchPresentKeys             PASS
=== TestSearchAbsentKeys              PASS
=== TestSearchStrings                 PASS

Inserción (6)
=== TestInsertSingleAndLen            PASS
=== TestInsertInOrderSorted           PASS
=== TestInsertDuplicatesIgnored       PASS
=== TestInsertAscendingAndDescending  PASS  (200 inserciones asc y desc)
=== TestInsertRandomizedStress        PASS  (2000 inserciones aleatorias)
=== TestInsertStringsBalanced         PASS

Consultas por rango (2)
=== TestRangeQuery                    PASS  (8 casos borde)
=== TestRangeQueryMatchesInOrderFilter PASS (100 rangos aleatorios)

Eliminación (8)
=== TestDeleteNonExistent             PASS
=== TestDeleteEmptyTree               PASS
=== TestDeleteOnlyKey                 PASS
=== TestDeleteFrom3NodeLeaf           PASS
=== TestDeleteCausingMergeAndHeightDecrease  PASS
=== TestDeleteInternalNode            PASS
=== TestDeleteAllKeysOneByOne         PASS  (borra las 15 claves en otro orden)
=== TestDeleteStressRandom            PASS  (500 inserciones, borra la mitad)

21/21 PASS
```

El helper `checkInvariants` verifica en cada test que:
1. Las claves en cada nodo están ordenadas.
2. Todas las hojas se encuentran al mismo nivel (árbol balanceado).

---

## Uso de Inteligencia Artificial

Este proyecto fue desarrollado con asistencia de **Claude** (Anthropic, modelo Sonnet 4.6 / Opus 4.8). El rol de la IA fue:

- Implementar la app SQLite (`cmd/dbapp/`) y la API REST (`cmd/api/`).
- Construir el frontend Vue 3 (`web/index.html`).
- Redactar este README y el análisis Big-O.

El rol de los integrantes fue:

- Definir los requisitos y validar cada entrega.
- Revisar y entender cada pieza del código (requisito de la rúbrica).
- Gestionar el repositorio Git (todos los commits son de autoría propia).
- Conseguir y preparar el dataset.
- Grabar el video y armar la presentación PPTX.
