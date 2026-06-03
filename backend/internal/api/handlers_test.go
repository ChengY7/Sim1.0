package api_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chengyang/sim1.0/backend/internal/api"
	fifaconfig "github.com/chengyang/sim1.0/backend/internal/fifa/config"
	nbaconfig "github.com/chengyang/sim1.0/backend/internal/nba/config"
)

var testRouter http.Handler

func TestMain(m *testing.M) {
	cfg, err := nbaconfig.Load()
	if err != nil {
		log.Fatalf("nbaconfig.Load: %v", err)
	}
	fifaCfg, err := fifaconfig.Load()
	if err != nil {
		log.Fatalf("fifaconfig.Load: %v", err)
	}
	testRouter = api.NewRouter(api.NewHandlers(cfg, fifaCfg), "http://localhost:3000")
	os.Exit(m.Run())
}

func simulate(t *testing.T, req api.SimulateRequest) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/nba/simulate", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(rec, r)
	return rec
}

// --- GET /nba/teams ---

func TestListTeams_OK(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nba/teams", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp api.ListTeamsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Teams) == 0 {
		t.Error("expected at least one team")
	}
	for i := 1; i < len(resp.Teams); i++ {
		if resp.Teams[i].ID < resp.Teams[i-1].ID {
			t.Errorf("teams not sorted by ID: %q before %q", resp.Teams[i-1].ID, resp.Teams[i].ID)
		}
	}
}

// --- POST /nba/simulate ---

func TestSimulate_Valid(t *testing.T) {
	seed := int64(42)
	rec := simulate(t, api.SimulateRequest{HomeTeamID: "LAL", AwayTeamID: "BOS", Seed: &seed})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp api.SimulateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.State.Status != "final" {
		t.Errorf("status = %q, want final", resp.State.Status)
	}
	if resp.Seed != seed {
		t.Errorf("seed = %d, want %d", resp.Seed, seed)
	}
	if resp.Truncated {
		t.Error("should not be truncated")
	}
	if len(resp.Events) == 0 {
		t.Error("expected events")
	}
}

func TestSimulate_MissingAwayTeam(t *testing.T) {
	rec := simulate(t, api.SimulateRequest{HomeTeamID: "LAL"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestSimulate_MissingHomeTeam(t *testing.T) {
	rec := simulate(t, api.SimulateRequest{AwayTeamID: "BOS"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestSimulate_SameTeam(t *testing.T) {
	rec := simulate(t, api.SimulateRequest{HomeTeamID: "LAL", AwayTeamID: "LAL"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestSimulate_UnknownTeam(t *testing.T) {
	rec := simulate(t, api.SimulateRequest{HomeTeamID: "LAL", AwayTeamID: "FAKE"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestSimulate_InvalidJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/nba/simulate", bytes.NewBufferString("{bad json}"))
	r.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(rec, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestSimulate_OversizedBody(t *testing.T) {
	body := bytes.Repeat([]byte("x"), 5000)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/nba/simulate", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	testRouter.ServeHTTP(rec, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCORSHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nba/teams", nil))
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "" {
		t.Error("expected Access-Control-Allow-Origin header")
	}
}

func TestCORSPreflight(t *testing.T) {
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodOptions, "/nba/simulate", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("OPTIONS status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "" {
		t.Error("expected Access-Control-Allow-Origin header on preflight")
	}
}

func TestSimulate_OT(t *testing.T) {
	seed := int64(80)
	rec := simulate(t, api.SimulateRequest{HomeTeamID: "LAL", AwayTeamID: "BOS", Seed: &seed})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp api.SimulateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.State.Period != "OT" {
		t.Errorf("period = %q, want OT", resp.State.Period)
	}
	if resp.State.HomeScore == resp.State.AwayScore {
		t.Error("OT game must not end tied")
	}
}

func TestSimulate_2OT(t *testing.T) {
	seed := int64(83)
	rec := simulate(t, api.SimulateRequest{HomeTeamID: "LAL", AwayTeamID: "BOS", Seed: &seed})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp api.SimulateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.State.Period != "2OT" {
		t.Errorf("period = %q, want 2OT", resp.State.Period)
	}
	if resp.State.HomeScore == resp.State.AwayScore {
		t.Error("2OT game must not end tied")
	}
}
