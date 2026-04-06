package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stockyard-dev/stockyard-harvest/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/crops", s.listCrops)
	s.mux.HandleFunc("POST /api/crops", s.createCrops)
	s.mux.HandleFunc("GET /api/crops/export.csv", s.exportCrops)
	s.mux.HandleFunc("GET /api/crops/{id}", s.getCrops)
	s.mux.HandleFunc("PUT /api/crops/{id}", s.updateCrops)
	s.mux.HandleFunc("DELETE /api/crops/{id}", s.delCrops)
	s.mux.HandleFunc("GET /api/harvests", s.listHarvests)
	s.mux.HandleFunc("POST /api/harvests", s.createHarvests)
	s.mux.HandleFunc("GET /api/harvests/export.csv", s.exportHarvests)
	s.mux.HandleFunc("GET /api/harvests/{id}", s.getHarvests)
	s.mux.HandleFunc("PUT /api/harvests/{id}", s.updateHarvests)
	s.mux.HandleFunc("DELETE /api/harvests/{id}", s.delHarvests)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", s.tierHandler)
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(c); json.NewEncoder(w).Encode(v) }
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", 302) }
func oe[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

func (s *Server) listCrops(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("status"); v != "" { filters["status"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"crops": oe(s.db.SearchCrops(q, filters))}); return }
	wj(w, 200, map[string]any{"crops": oe(s.db.ListCrops())})
}

func (s *Server) createCrops(w http.ResponseWriter, r *http.Request) {
	if s.limits.Tier == "none" { we(w, 402, "No license key. Start a 14-day trial at https://stockyard.dev/for/"); return }
	if s.limits.TrialExpired { we(w, 402, "Trial expired. Subscribe at https://stockyard.dev/pricing/"); return }
	var e store.Crops
	json.NewDecoder(r.Body).Decode(&e)
	if e.Name == "" { we(w, 400, "name required"); return }
	s.db.CreateCrops(&e)
	wj(w, 201, s.db.GetCrops(e.ID))
}

func (s *Server) getCrops(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetCrops(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateCrops(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetCrops(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Crops
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.Name == "" { patch.Name = existing.Name }
	if patch.Variety == "" { patch.Variety = existing.Variety }
	if patch.Field == "" { patch.Field = existing.Field }
	if patch.PlantedDate == "" { patch.PlantedDate = existing.PlantedDate }
	if patch.ExpectedHarvest == "" { patch.ExpectedHarvest = existing.ExpectedHarvest }
	if patch.Status == "" { patch.Status = existing.Status }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateCrops(&patch)
	wj(w, 200, s.db.GetCrops(patch.ID))
}

func (s *Server) delCrops(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteCrops(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportCrops(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListCrops()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=crops.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "name", "variety", "field", "planted_date", "expected_harvest", "status", "notes", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.Name), fmt.Sprintf("%v", e.Variety), fmt.Sprintf("%v", e.Field), fmt.Sprintf("%v", e.PlantedDate), fmt.Sprintf("%v", e.ExpectedHarvest), fmt.Sprintf("%v", e.Status), fmt.Sprintf("%v", e.Notes), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) listHarvests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("unit"); v != "" { filters["unit"] = v }
	if v := r.URL.Query().Get("quality"); v != "" { filters["quality"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"harvests": oe(s.db.SearchHarvests(q, filters))}); return }
	wj(w, 200, map[string]any{"harvests": oe(s.db.ListHarvests())})
}

func (s *Server) createHarvests(w http.ResponseWriter, r *http.Request) {
	var e store.Harvests
	json.NewDecoder(r.Body).Decode(&e)
	if e.CropId == "" { we(w, 400, "crop_id required"); return }
	if e.Date == "" { we(w, 400, "date required"); return }
	s.db.CreateHarvests(&e)
	wj(w, 201, s.db.GetHarvests(e.ID))
}

func (s *Server) getHarvests(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetHarvests(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateHarvests(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetHarvests(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Harvests
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.CropId == "" { patch.CropId = existing.CropId }
	if patch.Date == "" { patch.Date = existing.Date }
	if patch.Unit == "" { patch.Unit = existing.Unit }
	if patch.Quality == "" { patch.Quality = existing.Quality }
	if patch.Notes == "" { patch.Notes = existing.Notes }
	s.db.UpdateHarvests(&patch)
	wj(w, 200, s.db.GetHarvests(patch.ID))
}

func (s *Server) delHarvests(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteHarvests(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportHarvests(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListHarvests()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=harvests.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "crop_id", "date", "quantity", "unit", "quality", "notes", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.CropId), fmt.Sprintf("%v", e.Date), fmt.Sprintf("%v", e.Quantity), fmt.Sprintf("%v", e.Unit), fmt.Sprintf("%v", e.Quality), fmt.Sprintf("%v", e.Notes), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{}
	m["crops_total"] = s.db.CountCrops()
	m["harvests_total"] = s.db.CountHarvests()
	wj(w, 200, m)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"status": "ok", "service": "harvest"}
	m["crops"] = s.db.CountCrops()
	m["harvests"] = s.db.CountHarvests()
	wj(w, 200, m)
}

// loadPersonalConfig reads config.json from the data directory.
func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg
	log.Printf("loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}
