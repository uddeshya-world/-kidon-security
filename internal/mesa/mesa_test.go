package mesa

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEventIDIsStableAndFormatMatchesMesaD(t *testing.T) {
	at := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	a := ConnectEvent("shomer-host-1", "arn:aws:iam::111111111111:role/eval-c", "wiki.example.org", 42, at)
	b := ConnectEvent("shomer-host-1", "arn:aws:iam::111111111111:role/eval-c", "wiki.example.org", 42, at)
	if a.EventID != b.EventID || len(a.EventID) != 64 {
		t.Fatalf("unstable id")
	}
	raw, _ := json.Marshal(a)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	for _, k := range []string{"event_version", "source_id", "observed_at", "kind", "principal", "host", "event_id"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing %s", k)
		}
	}
	if m["kind"] != "http" || m["method"] != "CONNECT" {
		t.Fatalf("unexpected %v", m)
	}
}

func TestEmitterFlushesEachEvent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "events.jsonl")
	e, err := NewFileEmitter(p)
	if err != nil {
		t.Fatal(err)
	}
	_ = e.Emit(ExecEvent("shomer", "workload", 7, "bash", time.Unix(0, 0)))
	// Read before Close: the line must already be on disk.
	f, _ := os.Open(p)
	sc := bufio.NewScanner(f)
	if !sc.Scan() || !strings.Contains(sc.Text(), `"kind":"exec"`) {
		t.Fatalf("event not flushed")
	}
	f.Close()
	_ = e.Close()
}

const report = `{"version":"0.4.0","mode":"lattice","violation_count":1,"near_miss_count":1,"agents":[
 {"agent_id":"agent:b","zone":"production","closure":{"P":2,"U":2,"E":1},"threshold":{"P":2,"U":2,"E":2},"violation":false,"near_miss":true,"missing":["E"],"witness":null,"min_cut":null,"cut_dimension":null},
 {"agent_id":"agent:c","zone":"production","closure":{"P":3,"U":2,"E":3},"threshold":{"P":2,"U":2,"E":2},"violation":true,"near_miss":false,"missing":[],
  "witness":{"P":["svc:data","agent:c"],"U":["svc:tickets","agent:b","svc:cache","agent:c"],"E":["agent:c"]},
  "min_cut":[{"src":"svc:cache","dst":"agent:c","flow_type":"read"}],"cut_dimension":"U"}]}`

func TestSummaryListsInv01FirstWithWitnessAndCut(t *testing.T) {
	r, err := ParseReport([]byte(report))
	if err != nil {
		t.Fatal(err)
	}
	s := Summary(r)
	if !strings.HasPrefix(strings.SplitN(s, "\n\n", 2)[1], "agent:c") {
		t.Fatalf("INV01 not first:\n%s", s)
	}
	if !strings.Contains(s, "witness U: svc:tickets -> agent:b -> svc:cache -> agent:c") || !strings.Contains(s, "minimal cut: svc:cache -> agent:c (read)") {
		t.Fatalf("missing witness or cut:\n%s", s)
	}
}

func TestParseReportRejectsBooleanMode(t *testing.T) {
	if _, err := ParseReport([]byte(`{"version":"0.3.2","agents":[]}`)); err == nil {
		t.Fatal("expected error for non-lattice report")
	}
}

// labServer simulates a shared surface that forwards whatever is written to it to an egress observer.
func labServer(forward bool) *httptest.Server {
	var mu sync.Mutex
	seen := map[string]bool{}
	mux := http.NewServeMux()
	mux.HandleFunc("/surface/cache", func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 64)
		n, _ := r.Body.Read(b)
		if forward {
			mu.Lock()
			seen[string(b[:n])] = true
			mu.Unlock()
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/observer", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if tok := r.URL.Query().Get("token"); seen[tok] {
			_, _ = w.Write([]byte(tok))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(mux)
}

func finding() AgentFinding {
	r, _ := ParseReport([]byte(report))
	return r.Agents[1]
}

func TestSC12WitnessReplayDemonstratedInLab(t *testing.T) {
	srv := labServer(true)
	defer srv.Close()
	host := mustHost(srv.URL)
	lab := Lab{Surfaces: map[string]string{"svc:cache": srv.URL + "/surface/cache"}, Observer: srv.URL + "/observer",
		AllowedHosts: []string{host}, Wait: time.Second}
	r := lab.ReplayWitness(finding(), "U")
	if !r.Demonstrated || r.Planted != "svc:cache" {
		t.Fatalf("expected demonstrated: %+v", r)
	}
}

func TestSC12NotDemonstratedWhenCanaryNeverReachesEgress(t *testing.T) {
	srv := labServer(false)
	defer srv.Close()
	lab := Lab{Surfaces: map[string]string{"svc:cache": srv.URL + "/surface/cache"}, Observer: srv.URL + "/observer",
		AllowedHosts: []string{mustHost(srv.URL)}, Wait: 200 * time.Millisecond}
	r := lab.ReplayWitness(finding(), "U")
	if r.Demonstrated || !strings.Contains(r.Reason, "not demonstrated") {
		t.Fatalf("expected not demonstrated: %+v", r)
	}
}

func TestStrikeRefusesNonLabHosts(t *testing.T) {
	lab := Lab{Surfaces: map[string]string{"svc:cache": "https://prod.example.com/x"}, Observer: "http://127.0.0.1:1/observer",
		AllowedHosts: []string{"127.0.0.1"}, Wait: 0}
	r := lab.ReplayWitness(finding(), "U")
	if r.Demonstrated || !strings.Contains(r.Reason, "not an allowlisted lab host") {
		t.Fatalf("expected refusal: %+v", r)
	}
}

func mustHost(raw string) string {
	u, _ := url.Parse(raw)
	return u.Hostname()
}
