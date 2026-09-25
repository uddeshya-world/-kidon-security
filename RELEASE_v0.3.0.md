# ⚔️ Kidon v0.3.0: The "Command" Update
> **Agentic Cyber Defense Platform**

Kidon v0.3.0 introduces the **Command Cockpit (TUI)** and the **Supply Chain Intelligence** module (codenamed **"The Gatekeeper"**).

### 🚀 New Features

**1. The Command Cockpit (TUI)**
* **Unified Dashboard:** A cyberpunk, keyboard-driven interface (`./kidon dashboard`) powered by `Bubble Tea`.
* **Live Monitoring:** Watch real-time logs from the Shomer (Guard) and Sentry (Scanner) in one terminal.
* **Interactive Tabs:** Switch between Sentry, Shomer, and Kidon Strike modes instantly.

**2. Supply Chain Intelligence ("The Gatekeeper")**
* **OWASP ASI-04 Coverage:** The Gatekeeper module now deeply parses `requirements.txt` and `go.mod` files.
* **OSV.dev Integration:** Automatically cross-references your agent's dependencies against the Open Source Vulnerabilities (OSV) database to intercept poisoned or vulnerable packages before installation.

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
