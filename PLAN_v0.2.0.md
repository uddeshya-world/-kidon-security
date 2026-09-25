# ⚔️ Kidon v0.2.0: Operation Iron Dome (Network Fortress)

**Objective:** Evolve Kidon from a "Process Guard" to a "Network Fortress" with Supply Chain Intelligence.
**Philosophy:** "First Time Right" — Security features must be fail-safe by default.
**Tech Stack:** Go 1.23, Cilium eBPF, C, Docker.

---

## 1. Architectural Upgrades

### A. The Network Layer (Network Fortress)
We are moving from Process Security (`execve`) to Socket Security (`connect`).
* **Mechanism:** `cgroup` eBPF hooks. This allows us to filter traffic for the specific container without affecting the host node excessively.
* **Data Structure:** `BPF_MAP_TYPE_HASH` shared between Kernel and User Space.
* **Fail-Safe:** IPv6 is **blocked by default** to prevent bypass attacks.

### B. Supply Chain Intelligence
* **Mechanism:** Static parsing of dependency files (`requirements.txt`, `go.mod`) + OSV.dev API.
* **Integration:** Results are merged into the `mission_report.html`.

---

## 2. Implementation Phases

### Phase 4: The Network Fortress (Priority P0)
**Goal:** Prevent Data Exfiltration (OWASP ASI-02).

* **Task 4.1: The Kernel Module (`bpf/network_monitor.c`)**
    * **Hook 1:** `SEC("cgroup/connect4")`.
        * Look up destination IP in `allowed_ips` Hash Map.
        * If found: `return 1` (ALLOW).
        * If missing: `return 0` (BLOCK) & emit event to `network_events` RingBuffer.
    * **Hook 2:** `SEC("cgroup/connect6")`.
        * **Strict Rule:** Always `return 0` (BLOCK).
        * *Rationale:* We are not building IPv6 filtering logic in v0.2.0. To avoid bypass, we must disable it entirely for the agent.

* **Task 4.2: The Policy Engine (`internal/runtime/network_guard.go`)**
    * **Struct:** `NetworkGuard` holding the `ebpf.Map` reference.
    * **DNS Logic:**
        * Loop through domains in `kidon_policy.yaml`.
        * Resolve to A records (IPv4).
        * Update Map: `Map.Update(ip_int, 1, BPF_ANY)`.
    * **Ticker:** Run every **30 seconds** (Critical for dynamic Cloud IPs).
    * **Error Handling:** If Map is full, log error but *do not crash*.

* **Task 4.3: Integration & Alerting**
    * Update `StartGuard()` in `monitor.go` to attach the new network hooks.
    * Read from `network_events` RingBuffer.
    * Log: `🚨 NETWORK BLOCK: Process 'python' tried to connect to 1.2.3.4 (Not Allowed)`.

### Phase 5: Supply Chain Scanner (Priority P1)
**Goal:** Detect Vulnerable Dependencies (OWASP ASI-04).

* **Task 5.1: Dependency Parser (`internal/static/parser.go`)**
    * Support `requirements.txt` (Python) and `go.mod` (Go).
    * Extract `{name, version}` pairs.

* **Task 5.2: OSV Client (`internal/static/osv.go`)**
    * Batch query `https://api.osv.dev/v1/querybatch`.
    * Map responses to Kidon `Issue` structs.

### Phase 6: Production Hardening (Priority P2)
**Goal:** CI/CD Stability.

* **Task 6.1: GitHub Action (`action.yml`)**
    * Define a Docker container action.
    * **Crucial:** Must run with `--privileged` and `--cgroupns=host`.

---

## 3. Risk Mitigation Strategy (The "Manager's Checklist")

| Risk | Mitigation Implemented |
| :--- | :--- |
| **IPv6 Bypass** | **Task 4.1:** Explicitly added `cgroup/connect6` hook that returns `0` (Block). |
| **DNS Staleness** | **Task 4.2:** DNS Refresh rate set to 30s (high frequency) to minimize false positives. |
| **Map Overflow** | **Task 4.2:** Explicit error handling for `Map.Update`. |
| **Windows Dev** | **Stub Files:** Ensure `network_guard_stub.go` exists so the binary compiles on non-Linux machines (even if features are disabled). |

---

## 4. Verification Plan (QA)

### Test Case C1: The "OpenAI" Test
* **Config:** Allow `api.openai.com`.
* **Action:** Agent runs `curl https://api.openai.com/v1/models`.
* **Expectation:** **PASS** (200 OK).

### Test Case C2: The "Exfiltration" Test
* **Config:** Allow `api.openai.com`.
* **Action:** Agent runs `curl http://evil-server.com`.
* **Expectation:** **BLOCK** (Connection Refused).
* **Log Check:** CLI shows "Blocked connection to [Evil IP]".

### Test Case C3: The "IPv6" Test
* **Action:** Agent runs `curl -6 google.com`.
* **Expectation:** **BLOCK** (Immediate failure).
