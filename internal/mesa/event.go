// Package mesa connects Kidon to MESA formats. Dependencies run one way: Kidon consumes MESA's
// formats (observed-edge events, scanner reports); MESA depends on nothing in Kidon.
package mesa

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// EventVersion is the ACM observed-edge event format version (mesa-ibi-scanner docs/MESA-D-DESIGN.md §3).
const EventVersion = "0.1.0"

// Event is one observed edge in the ACM event format that mesa-d ingests (`--acm`).
type Event struct {
	EventVersion string `json:"event_version"`
	SourceID     string `json:"source_id"`
	ObservedAt   string `json:"observed_at"`
	Kind         string `json:"kind"`
	Principal    string `json:"principal"`
	Action       string `json:"action"`
	Resource     string `json:"resource"`
	Method       string `json:"method"`
	Host         string `json:"host"`
	Path         string `json:"path"`
	BytesOut     int64  `json:"bytes_out"`
	SessionID    string `json:"session_id"`
	Token        string `json:"token"`
	RawRef       string `json:"raw_ref"`
	EventID      string `json:"event_id"`
}

// WithID fills EventID with the SHA-256 of the event's other fields.
func (e Event) WithID() Event {
	e.EventVersion = EventVersion
	e.EventID = ""
	b, _ := json.Marshal(e)
	sum := sha256.Sum256(b)
	e.EventID = hex.EncodeToString(sum[:])
	return e
}

// ExecEvent converts a Shomer process-execution record. The principal is the workload identity the
// host belongs to (set by the operator), never a value the agent supplies.
func ExecEvent(sourceID, workload string, pid uint32, comm string, at time.Time) Event {
	return Event{
		SourceID:   sourceID,
		ObservedAt: at.UTC().Format(time.RFC3339Nano),
		Kind:       "exec",
		Principal:  workload,
		Action:     "process:exec",
		Resource:   comm,
		RawRef:     fmt.Sprintf("shomer:exec#pid=%d", pid),
	}.WithID()
}

// ConnectEvent converts a Shomer egress connect record into an http-kind edge (host only; no payload).
func ConnectEvent(sourceID, workload, host string, pid uint32, at time.Time) Event {
	return Event{
		SourceID:   sourceID,
		ObservedAt: at.UTC().Format(time.RFC3339Nano),
		Kind:       "http",
		Principal:  workload,
		Method:     "CONNECT",
		Host:       host,
		RawRef:     fmt.Sprintf("shomer:connect#pid=%d", pid),
	}.WithID()
}

// Emitter writes events as JSON lines and flushes after each one, so a sink on another host receives them
// immediately. An agent that reaches host root could tamper with local eBPF; shipping off-host limits that.
type Emitter struct {
	mu sync.Mutex
	w  *bufio.Writer
	f  *os.File
}

// NewFileEmitter appends to path (typically a pipe or a file tailed by a shipper to the log account).
func NewFileEmitter(path string) (*Emitter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	return &Emitter{w: bufio.NewWriter(f), f: f}, nil
}

// Emit writes one event and flushes.
func (e *Emitter) Emit(ev Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	b, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	if _, err := e.w.Write(append(b, '\n')); err != nil {
		return err
	}
	return e.w.Flush()
}

// Close flushes and closes the sink.
func (e *Emitter) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	_ = e.w.Flush()
	return e.f.Close()
}
