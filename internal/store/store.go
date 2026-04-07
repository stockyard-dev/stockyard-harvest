package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Crops struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Variety string `json:"variety"`
	Field string `json:"field"`
	PlantedDate string `json:"planted_date"`
	ExpectedHarvest string `json:"expected_harvest"`
	Status string `json:"status"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

type Harvests struct {
	ID string `json:"id"`
	CropId string `json:"crop_id"`
	Date string `json:"date"`
	Quantity float64 `json:"quantity"`
	Unit string `json:"unit"`
	Quality string `json:"quality"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil { return nil, err }
	db, err := sql.Open("sqlite", filepath.Join(d, "harvest.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE IF NOT EXISTS crops(id TEXT PRIMARY KEY, name TEXT NOT NULL, variety TEXT DEFAULT '', field TEXT DEFAULT '', planted_date TEXT DEFAULT '', expected_harvest TEXT DEFAULT '', status TEXT DEFAULT '', notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS harvests(id TEXT PRIMARY KEY, crop_id TEXT NOT NULL, date TEXT NOT NULL, quantity REAL DEFAULT 0, unit TEXT DEFAULT '', quality TEXT DEFAULT '', notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(resource TEXT NOT NULL, record_id TEXT NOT NULL, data TEXT NOT NULL DEFAULT '{}', PRIMARY KEY(resource, record_id))`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) CreateCrops(e *Crops) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO crops(id, name, variety, field, planted_date, expected_harvest, status, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.Name, e.Variety, e.Field, e.PlantedDate, e.ExpectedHarvest, e.Status, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetCrops(id string) *Crops {
	var e Crops
	if d.db.QueryRow(`SELECT id, name, variety, field, planted_date, expected_harvest, status, notes, created_at FROM crops WHERE id=?`, id).Scan(&e.ID, &e.Name, &e.Variety, &e.Field, &e.PlantedDate, &e.ExpectedHarvest, &e.Status, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListCrops() []Crops {
	rows, _ := d.db.Query(`SELECT id, name, variety, field, planted_date, expected_harvest, status, notes, created_at FROM crops ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Crops
	for rows.Next() { var e Crops; rows.Scan(&e.ID, &e.Name, &e.Variety, &e.Field, &e.PlantedDate, &e.ExpectedHarvest, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateCrops(e *Crops) error {
	_, err := d.db.Exec(`UPDATE crops SET name=?, variety=?, field=?, planted_date=?, expected_harvest=?, status=?, notes=? WHERE id=?`, e.Name, e.Variety, e.Field, e.PlantedDate, e.ExpectedHarvest, e.Status, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteCrops(id string) error {
	_, err := d.db.Exec(`DELETE FROM crops WHERE id=?`, id)
	return err
}

func (d *DB) CountCrops() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM crops`).Scan(&n); return n
}

func (d *DB) SearchCrops(q string, filters map[string]string) []Crops {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (name LIKE ? OR variety LIKE ? OR field LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["status"]; ok && v != "" { where += " AND status=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, name, variety, field, planted_date, expected_harvest, status, notes, created_at FROM crops WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Crops
	for rows.Next() { var e Crops; rows.Scan(&e.ID, &e.Name, &e.Variety, &e.Field, &e.PlantedDate, &e.ExpectedHarvest, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) CreateHarvests(e *Harvests) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO harvests(id, crop_id, date, quantity, unit, quality, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.CropId, e.Date, e.Quantity, e.Unit, e.Quality, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetHarvests(id string) *Harvests {
	var e Harvests
	if d.db.QueryRow(`SELECT id, crop_id, date, quantity, unit, quality, notes, created_at FROM harvests WHERE id=?`, id).Scan(&e.ID, &e.CropId, &e.Date, &e.Quantity, &e.Unit, &e.Quality, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListHarvests() []Harvests {
	rows, _ := d.db.Query(`SELECT id, crop_id, date, quantity, unit, quality, notes, created_at FROM harvests ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Harvests
	for rows.Next() { var e Harvests; rows.Scan(&e.ID, &e.CropId, &e.Date, &e.Quantity, &e.Unit, &e.Quality, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateHarvests(e *Harvests) error {
	_, err := d.db.Exec(`UPDATE harvests SET crop_id=?, date=?, quantity=?, unit=?, quality=?, notes=? WHERE id=?`, e.CropId, e.Date, e.Quantity, e.Unit, e.Quality, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteHarvests(id string) error {
	_, err := d.db.Exec(`DELETE FROM harvests WHERE id=?`, id)
	return err
}

func (d *DB) CountHarvests() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM harvests`).Scan(&n); return n
}

func (d *DB) SearchHarvests(q string, filters map[string]string) []Harvests {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (crop_id LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["unit"]; ok && v != "" { where += " AND unit=?"; args = append(args, v) }
	if v, ok := filters["quality"]; ok && v != "" { where += " AND quality=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, crop_id, date, quantity, unit, quality, notes, created_at FROM harvests WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Harvests
	for rows.Next() { var e Harvests; rows.Scan(&e.ID, &e.CropId, &e.Date, &e.Quantity, &e.Unit, &e.Quality, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

// GetExtras returns the JSON extras blob for a record. Returns "{}" if none.
func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(`SELECT data FROM extras WHERE resource=? AND record_id=?`, resource, recordID).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

// SetExtras stores the JSON extras blob for a record.
func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?) ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`, resource, recordID, data)
	return err
}

// DeleteExtras removes extras when a record is deleted.
func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(`DELETE FROM extras WHERE resource=? AND record_id=?`, resource, recordID)
	return err
}

// AllExtras returns all extras for a resource type as a map of record_id → JSON string.
func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(`SELECT record_id, data FROM extras WHERE resource=?`, resource)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
