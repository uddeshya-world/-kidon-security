# ⚔️ KIDON (כידון)
> **Agentic Cyber Defense Platform**

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tech: Cilium eBPF](https://img.shields.io/badge/Powered%20By-Cilium%20eBPF-F6C702)](https://ebpf.io)

**Kidon** is a security platform for autonomous AI agents. It provides a unified **Command Cockpit (TUI)** combining Static Analysis, Runtime eBPF Guarding, and Red Teaming.

![Kidon Command Cockpit - The Sentry](assets/dashboard_sentry.png)


---

## 🚀 Quick Start 

### Option 1: Docker image (not yet available)

No image is published. An anonymous pull of `ghcr.io/uddeshya-world/kidon:latest` returns 403.

```bash
# Run the Command Cockpit (TUI)
docker run -it --privileged --pid=host \
  -v $(pwd):/target \
  ghcr.io/uddeshya-world/kidon:latest dashboard

# Scan current directory for vulnerabilities
docker run --rm -v $(pwd):/target \
  ghcr.io/uddeshya-world/kidon:latest scan /target
```

### Option 2: Prebuilt binary (not yet available)

This repository has no releases, so there is no prebuilt binary to download.

```bash
curl -sfL https://raw.githubusercontent.com/uddeshya-world/-kidon-security/main/install.sh | sh
./kidon dashboard
```

### Option 3: Build from Source

On Linux, `internal/runtime/monitor.go` calls `bpfObjects` and `loadBpfObjects`. Those identifiers are emitted by bpf2go and are absent until you generate them, so `go build` fails if you skip that step. `gen.go` is tagged `linux`, and `go generate` runs:

```text
go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 -type event bpf ../../bpf/process_monitor.c -- -I../../bpf/headers
```

bpf2go comes from `github.com/cilium/ebpf` (already required by `go.mod`); `go run` builds it. The directive targets amd64 and writes `internal/runtime/bpf_x86_bpfel.go` plus the embedded object, compiled for `386` and `amd64`. A Linux arm64 build still misses those symbols after this command.

Generate needs:

- Go 1.24 or newer (see `go.mod`)
- `clang`, the compiler bpf2go invokes
- `llvm`, which provides `llvm-strip`
- `libbpf-dev`, which provides `<bpf/bpf_helpers.h>` and `<bpf/bpf_tracing.h>`
- `gcc-multilib`, which provides the `/usr/include/asm` symlink clang needs to resolve `<asm/types.h>` while compiling `<linux/bpf.h>`

On Debian or Ubuntu, from a fresh clone:

```bash
git clone https://github.com/uddeshya-world/-kidon-security
cd -kidon-security
sudo apt-get update
sudo apt-get install -y clang llvm libbpf-dev gcc-multilib
go generate ./internal/runtime/...
go build -o kidon cmd/kidon/main.go
./kidon dashboard
```

On other operating systems the Linux file is excluded and `go build -o kidon cmd/kidon/main.go` uses `internal/runtime/monitor_stub.go`.

---

## 🖥️ The Command Cockpit

A cyberpunk, keyboard-driven TUI that unifies all security operations.

| Tab | Engine | Controls |
|-----|--------|----------|
| **THE SENTRY** | demo output (static) | `[s]` Scan |
| **THE SHOMER** | demo output (simulated events) | `[c]` Clear `[p]` Pause |
| **THE KIDON** | demo output (static) | `[a]` Probe `[d]` DAN `[f]` Flood |

The TUI's Sentry `[s]` and Kidon `[a]`/`[d]`/`[f]`/`[e]` actions currently display fixed demonstration output and do not scan or contact a target. The Shomer tab shows simulated demo events. `kidon scan` performs the real scan, and `kidon guard` runs the eBPF guard.

![The Shomer tab (demo output)](assets/dashboard_shomer.png)

![The Kidon - Red Team](assets/dashboard_kidon.png)

---

## 📦 Supply Chain Intelligence ("The Gatekeeper")

`kidon scan` checks dependency files against OSV.dev.

```bash
./kidon scan ./my-agent-repo
```

**Supported:** `requirements.txt` | `go.mod` | `package.json`

Example output only. This block was not produced by `kidon scan` on this repository:

```text
⚔️  KIDON STATIC SCANNER
📦 Analyzing supply chain dependencies...
   ⚠ Found 74 vulnerable dependencies!

[CRITICAL] requests@2.0.0 - CVE-2018-18074
[CRITICAL] django@1.11.0 - 36 vulnerabilities!
```

---

## 🛡️ Capabilities

| Module | Code Name | Function |
| :--- | :--- | :--- |
| **Scanner** | *The Sentry* | Supply Chain + Secrets |
| **Guard** | *The Shomer* | eBPF Runtime Protection |
| **Strike** | *The Kidon* | AI-Powered Red Teaming |

---
<img width="1671" height="758" alt="image" src="https://github.com/user-attachments/assets/0fc35fa3-e0ff-42bd-af13-cbf4ad4258bd" />

## 📝 License

MIT License - See [LICENSE](LICENSE) for details.

---

*Built for the Age of Agentic AI.*
