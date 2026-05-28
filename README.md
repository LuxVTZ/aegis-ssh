<div align="center">
  <h1>🛡️ Aegis SSH Platform</h1>
  <p><b>Enterprise-grade MCP Server for SSH Fleet Orchestration</b></p>
  
  [![Go Version](https://img.shields.io/github/go-mod/go-version/LuxVTZ/aegis-ssh)](https://golang.org)
  [![License](https://img.shields.io/github/license/LuxVTZ/aegis-ssh)](https://github.com/LuxVTZ/aegis-ssh/blob/main/LICENSE)
  [![GitHub Release](https://img.shields.io/github/v/release/LuxVTZ/aegis-ssh)](https://github.com/LuxVTZ/aegis-ssh/releases)
  [![NPM Version](https://img.shields.io/npm/v/aegis-ssh-mcp)](https://www.npmjs.com/package/aegis-ssh-mcp)
</div>

<br>

Aegis is an ultra-fast, zero-dependency MCP (Model Context Protocol) Server written in Go. It transforms any AI assistant (like Claude, Cursor, Windsurf) into a powerful DevOps engineer capable of securely orchestrating fleets of VPS servers via SSH.

## 🌟 Why Aegis? (Aegis vs Python/Node MCPs)
Unlike typical Python (`uvx`) or NodeJS MCP servers that require environments, dependencies, and complex setups, **Aegis is a single, lightning-fast binary**.
- **No Dependencies:** Just download the binary and run it. No `pip`, `node_modules`, or `uvx` required.
- **Web Dashboard:** Type `aegis-ssh web` to launch a stunning Linear-style UI locally to monitor your servers and audit logs.
- **Mass Execution:** Powered by Go Goroutines, Aegis executes commands on dozens of servers concurrently.
- **Auto-Remediation:** Background daemon monitors your servers' CPU Load and dynamically alerts the AI if intervention is needed.
- **Reverse Tunneling:** Instantly expose local web servers to the public internet via your VPS using the `tunnel_expose` MCP tool.
- **Playbook Engine:** Run declarative YAML manifests to setup environments idempotently.

---

## 🚀 Installation

### Option 1: Direct Download (Recommended)
Download the pre-compiled binary for your OS (Windows, macOS, Linux) from the [GitHub Releases](https://github.com/LuxVTZ/aegis-ssh/releases) page.
Move it to a directory in your PATH.

### Option 2: Go Install
If you have Go 1.21+ installed on your machine:
```bash
go install github.com/LuxVTZ/aegis-ssh/cmd/aegis-ssh@latest
```

---

## 🔌 Connecting to AI Clients (MCP Setup)

We built an interactive wizard to automatically wire up Aegis to your favorite AI clients (Claude Desktop, Cursor). You don't even need to touch JSON files!

Just run the following command in your terminal:
```bash
npx -y aegis-ssh-mcp setup
```
**What this does:**
1. Prompts you to select your AI agents (Claude, Cursor).
2. Automatically locates and updates their configuration files to run the Aegis Zero-Install server.
3. Generates an `.aegis_skills.md` file in your directory with recommended system prompts for your AI.

If you prefer manual installation, use this snippet in your MCP config:
```json
{
  "mcpServers": {
    "aegis-ssh": {
      "command": "npx",
      "args": ["-y", "aegis-ssh-mcp", "server"]
    }
  }
}
```

---

## 💻 CLI Usage & Control Panel

Aegis isn't just an MCP server, it's a full developer tool.

```bash
# List all your servers (merged from ~/.ssh/config and ~/.sshmcp/machines.json)
aegis-ssh list

# Launch the Linear-style Web Dashboard on http://localhost:8080
aegis-ssh web --port 8080
```

## 🔒 Security & Audit
- Aegis supports strict Regex **Whitelisting** and **Blacklisting**.
- **SQLite Audit DB:** Every single command the AI executes is logged locally into `~/.sshmcp/audit.db`. You can view these logs directly in the Web Dashboard.
