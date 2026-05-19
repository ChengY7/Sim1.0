package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chengyang/sim1.0/backend/internal/api"
	"github.com/chengyang/sim1.0/backend/internal/config"
)

func newRouter(t *testing.T) http.Handler {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	return api.NewRouter(api.NewHandlers(cfg), "http://localhost:3000")
}

func simulate(t *testing.T, req api.SimulateRequest) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(req)
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/simulate", bytes.NewReader(body))
	newRouter(t).ServeHTTP(rec, r)
	return rec
}

// --- GET /teams ---

func TestListTeams_OK(t *testing.T) {
	rec := httptest.NewRecorder()
	newRouter(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/teams", nil))

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
		if resp.Teams[i].Name < resp.Teams[i-1].Name {
			t.Errorf("teams not sorted: %q before %q", resp.Teams[i-1].Name, resp.Teams[i].Name)
		}
	}
}

// --- POST /simulate ---

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
	r := httptest.NewRequest(http.MethodPost, "/simulate", bytes.NewBufferString("{bad json}"))
	newRouter(t).ServeHTTP(rec, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestSimulate_OT(t *testing.T) {
	seed := int64(16)
	rec := simulate(t, api.SimulateRequest{HomeTeamID: "LAL", AwayTeamID: "BOS", Seed: &seed})

	var resp api.SimulateResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.State.Period != "OT" {
		t.Errorf("period = %q, want OT", resp.State.Period)
	}
	if resp.State.HomeScore == resp.State.AwayScore {
		t.Error("OT game must not end tied")
	}
}

func TestSimulate_2OT(t *testing.T) {
	seed := int64(241)
	rec := simulate(t, api.SimulateRequest{HomeTeamID: "LAL", AwayTeamID: "BOS", Seed: &seed})

	var resp api.SimulateResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.State.Period != "2OT" {
		t.Errorf("period = %q, want 2OT", resp.State.Period)
	}
	if resp.State.HomeScore == resp.State.AwayScore {
		t.Error("2OT game must not end tied")
	}
}
