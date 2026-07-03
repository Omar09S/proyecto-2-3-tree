package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Omar09S/proyecto-2-3-tree/tree23"
	_ "modernc.org/sqlite"
)

const (
	csvPath = "data/ciudades.csv"
	dbPath  = "data/ciudades.db"
)

type Ciudad struct {
	Nombre    string
	Poblacion int
}

func main() {
	ciudades, err := leerCSV(csvPath)
	if err != nil {
		log.Fatalf("error leyendo CSV: %v", err)
	}
	fmt.Printf("CSV cargado: %d ciudades\n\n", len(ciudades))

	db, err := inicializarDB(dbPath)
	if err != nil {
		log.Fatalf("error abriendo BD: %v", err)
	}
	defer db.Close()

	if err := insertarCiudades(db, ciudades); err != nil {
		log.Fatalf("error insertando ciudades: %v", err)
	}

	arbol := tree23.New[string]()
	rows, err := db.Query("SELECT nombre FROM ciudades ORDER BY ROWID")
	if err != nil {
		log.Fatalf("error leyendo BD: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var nombre string
		if err := rows.Scan(&nombre); err != nil {
			log.Fatalf("error scaneando fila: %v", err)
		}
		arbol.Insert(nombre)
	}
	fmt.Printf("Árbol 2-3 construido con %d claves\n\n", arbol.Len())

	demoConsultas(arbol, db)
}

func demoConsultas(arbol *tree23.Tree23[string], db *sql.DB) {
	separador := func() { fmt.Println("────────────────────────────────────────") }

	separador()
	fmt.Println("a) BÚSQUEDA EXACTA")
	for _, nombre := range []string{"Lima", "Bogota", "Atlantida"} {
		encontrado := arbol.Search(nombre)
		fmt.Printf("   Search(%q) → %v", nombre, encontrado)
		if encontrado {
			pob := poblaciónDesdeBD(db, nombre)
			fmt.Printf("  (población: %s)", formatNum(pob))
		}
		fmt.Println()
	}

	separador()
	fmt.Println("b) RANGO ALFABÉTICO: ciudades entre \"C\" y \"M\"")
	rango := arbol.RangeQuery("C", "M￿") // ￿ asegura que incluye todo "M..."
	fmt.Printf("   %d ciudades encontradas:\n", len(rango))
	for _, nombre := range rango {
		pob := poblaciónDesdeBD(db, nombre)
		fmt.Printf("   %-30s %s hab.\n", nombre, formatNum(pob))
	}

	separador()
	fmt.Println("c) LISTADO COMPLETO (in-order ascendente)")
	todas := arbol.InOrder()
	for i, nombre := range todas {
		pob := poblaciónDesdeBD(db, nombre)
		fmt.Printf("   %2d. %-30s %s hab.\n", i+1, nombre, formatNum(pob))
	}
	separador()
	fmt.Printf("Total: %d ciudades únicas en el árbol.\n", arbol.Len())
}

func leerCSV(path string) ([]Ciudad, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var ciudades []Ciudad
	for _, rec := range records[1:] { // saltar cabecera
		if len(rec) < 2 {
			continue
		}
		pob, _ := strconv.Atoi(rec[1])
		ciudades = append(ciudades, Ciudad{Nombre: rec[0], Poblacion: pob})
	}
	return ciudades, nil
}

func inicializarDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ciudades (
		nombre    TEXT PRIMARY KEY,
		poblacion INTEGER NOT NULL
	)`)
	return db, err
}

func insertarCiudades(db *sql.DB, ciudades []Ciudad) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO ciudades (nombre, poblacion) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, c := range ciudades {
		if _, err := stmt.Exec(c.Nombre, c.Poblacion); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func poblaciónDesdeBD(db *sql.DB, nombre string) int {
	var pob int
	db.QueryRow("SELECT poblacion FROM ciudades WHERE nombre = ?", nombre).Scan(&pob)
	return pob
}

func formatNum(n int) string {
	s := strconv.Itoa(n)
	out := []byte{}
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}
