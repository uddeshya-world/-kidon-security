package mesa

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Lab maps MESA vertex ids to lab endpoints. Only hosts in AllowedHosts are ever contacted; Strike refuses
// anything else, so a witness replay cannot touch production or third-party systems.
type Lab struct {
	// Surfaces maps a vertex id on the witness path to a lab URL that accepts a canary via PUT.
	Surfaces map[string]string
	// Observer is the lab URL that reports whether a canary token reached egress (GET <observer>?token=...).
	Observer     string
	AllowedHosts []string
	Client       *http.Client
	// Wait is how long to poll the observer.
	Wait time.Duration
}

// Replay is the SC-12 outcome for one witness path.
type Replay struct {
	Agent        string   `json:"agent"`
	Dimension    string   `json:"dimension"`
	Path         []string `json:"path"`
	Token        string   `json:"token"`
	Planted      string   `json:"planted_on"`
	Demonstrated bool     `json:"demonstrated"`
	Reason       string   `json:"reason"`
}

func (l Lab) allowed(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	for _, h := range l.AllowedHosts {
		if strings.EqualFold(u.Hostname(), h) {
			return nil
		}
	}
	return fmt.Errorf("host %q is not an allowlisted lab host", u.Hostname())
}

func token() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "kidon-strike-" + hex.EncodeToString(b)
}

// ReplayWitness plants a unique canary on the first lab surface of the witness path for `dimension` and
// checks whether it reaches the lab egress observer. A finding the lab cannot reproduce is reported as
// not demonstrated; the finding itself stands.
func (l Lab) ReplayWitness(f AgentFinding, dimension string) Replay {
	path := f.Witness[dimension]
	r := Replay{Agent: f.AgentID, Dimension: dimension, Path: path}
	if !f.Violation || len(path) == 0 {
		r.Reason = "no INV01 witness for this dimension"
		return r
	}
	if err := l.allowed(l.Observer); err != nil {
		r.Reason = err.Error()
		return r
	}
	for _, v := range path {
		target, ok := l.Surfaces[v]
		if !ok {
			continue
		}
		if err := l.allowed(target); err != nil {
			r.Reason = err.Error()
			return r
		}
		r.Token = token()
		req, _ := http.NewRequest(http.MethodPut, target, strings.NewReader(r.Token))
		resp, err := l.client().Do(req)
		if err != nil || resp.StatusCode >= 300 {
			r.Reason = fmt.Sprintf("could not plant canary on %s", v)
			return r
		}
		resp.Body.Close()
		r.Planted = v
		break
	}
	if r.Planted == "" {
		r.Reason = "no lab surface mapped on the witness path"
		return r
	}
	deadline := time.Now().Add(l.Wait)
	for {
		resp, err := l.client().Get(l.Observer + "?token=" + url.QueryEscape(r.Token))
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && strings.Contains(string(body), r.Token) {
				r.Demonstrated = true
				r.Reason = "canary reached egress"
				return r
			}
		}
		if time.Now().After(deadline) {
			r.Reason = "canary not observed at egress within the wait window: finding marked not demonstrated"
			return r
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (l Lab) client() *http.Client {
	if l.Client != nil {
		return l.Client
	}
	return &http.Client{Timeout: 5 * time.Second}
}
