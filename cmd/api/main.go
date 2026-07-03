package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/Omar09S/proyecto-2-3-tree/tree23"
)

var (
	mu sync.RWMutex
	tr = tree23.New[int]()
)

func main() {
	preload([]int{15, 30, 45, 7, 22, 38, 52, 10, 18, 27, 35, 42, 50})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/tree", withCORS(handleTree))
	mux.HandleFunc("/api/insert", withCORS(handleInsert))
	mux.HandleFunc("/api/delete", withCORS(handleDelete))
	mux.HandleFunc("/api/search", withCORS(handleSearch))
	mux.HandleFunc("/api/range", withCORS(handleRange))
	mux.HandleFunc("/api/reset", withCORS(handleReset))
	mux.HandleFunc("/api/load", withCORS(handleLoad))
	mux.Handle("/", http.FileServer(http.Dir("web")))

	fmt.Println("Servidor en http://localhost:8080")
	fmt.Println("  Frontend: http://localhost:8080")
	fmt.Println("  API:      http://localhost:8080/api/tree")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleTree(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	snap := tr.Snapshot()
	size := tr.Len()
	mu.RUnlock()
	writeJSON(w, map[string]any{"tree": snap, "size": size})
}

func handleInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Key *int `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == nil {
		http.Error(w, `body inválido: se espera {"key": número}`, http.StatusBadRequest)
		return
	}

	mu.Lock()
	steps := tr.InsertSteps(*body.Key)
	snap := tr.Snapshot()
	size := tr.Len()
	mu.Unlock()

	inserted := len(steps) > 0 && steps[len(steps)-1].Phase != "duplicate"
	grew := false
	for _, s := range steps {
		if s.Phase == "grow" {
			grew = true
		}
	}

	writeJSON(w, map[string]any{
		"inserted": inserted,
		"grew":     grew,
		"steps":    steps,
		"tree":     snap,
		"size":     size,
	})
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Key *int `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == nil {
		http.Error(w, `body inválido: se espera {"key": número}`, http.StatusBadRequest)
		return
	}

	mu.Lock()
	steps := tr.DeleteSteps(*body.Key)
	snap := tr.Snapshot()
	size := tr.Len()
	mu.Unlock()

	deleted := len(steps) > 0 && steps[len(steps)-1].Phase != "not_found"

	writeJSON(w, map[string]any{
		"deleted": deleted,
		"steps":   steps,
		"tree":    snap,
		"size":    size,
	})
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Key *int `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == nil {
		http.Error(w, `body inválido: se espera {"key": número}`, http.StatusBadRequest)
		return
	}

	mu.RLock()
	found, steps := tr.SearchSteps(*body.Key)
	snap := tr.Snapshot()
	size := tr.Len()
	mu.RUnlock()

	writeJSON(w, map[string]any{
		"found": found,
		"steps": steps,
		"tree":  snap,
		"size":  size,
	})
}

func handleRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Lo *int `json:"lo"`
		Hi *int `json:"hi"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Lo == nil || body.Hi == nil {
		http.Error(w, `body inválido: se espera {"lo": número, "hi": número}`, http.StatusBadRequest)
		return
	}

	mu.RLock()
	results := tr.RangeQuery(*body.Lo, *body.Hi)
	snap := tr.Snapshot()
	size := tr.Len()
	mu.RUnlock()

	writeJSON(w, map[string]any{
		"results": results,
		"count":   len(results),
		"tree":    snap,
		"size":    size,
	})
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	mu.Lock()
	tr = tree23.New[int]()
	mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "size": 0, "tree": nil})
}

func handleLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	seen := make(map[int]bool)
	var nums []int
	for len(nums) < 20 {
		n := rand.Intn(200) + 1
		if !seen[n] {
			seen[n] = true
			nums = append(nums, n)
		}
	}

	mu.Lock()
	tr = tree23.New[int]()
	for _, n := range nums {
		tr.Insert(n)
	}
	snap := tr.Snapshot()
	size := tr.Len()
	mu.Unlock()

	writeJSON(w, map[string]any{"ok": true, "loaded": size, "tree": snap, "size": size})
}

func preload(keys []int) {
	for _, k := range keys {
		tr.Insert(k)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func withCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}
