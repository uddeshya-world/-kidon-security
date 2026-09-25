# ⚔️ Kidon v0.3.0: The "Command" Update
> **Agentic Cyber Defense Platform**

Kidon v0.3.0 introduces the **Command Cockpit (TUI)** and the **Supply Chain Intelligence** module (codenamed **"The Gatekeeper"**).

### 🚀 New Features

**1. The Command Cockpit (TUI)**
* **Unified Dashboard:** A cyberpunk, keyboard-driven interface (`./kidon dashboard`) powered by `Bubble Tea`.
* **Live Monitoring:** Watch real-time logs from the Shomer (Guard) in one terminal. The Sentry tab shows fixed demonstration output and does not scan or contact a target. `kidon scan` performs the real scan.
* **Interactive Tabs:** Switch between Sentry, Shomer, and Kidon Strike modes instantly.

**2. Supply Chain Intelligence ("The Gatekeeper")**
* The Gatekeeper module parses dependency manifests (`requirements.txt`, `go.mod`, and `package.json`) and flags packages that OSV.dev reports as vulnerable.
* **OSV.dev Integration:** Automatically cross-references your agent's dependencies against the Open Source Vulnerabilities (OSV) database to flag poisoned or vulnerable packages in dependency manifests.

**3. Network Fortress (Experimental)**
* **Passive DNS Mode:** Resolves agent domains to IPs for audit logging (Blocking is currently disabled for compatibility).

### 🛠️ Usage

**Run the Dashboard:**
```bash
./kidon dashboard
```

**Run the Supply Chain Scan:**
```bash
./kidon scan ./my-agent-repo
```

### 📦 Installation
```bash
git clone https://github.com/uddeshya-world/-kidon-security
cd -kidon-security
go build -o kidon cmd/kidon/main.go
```

---

*Built for the Age of Agentic AI.*
