const fs = require('fs');
const path = require('path');
const os = require('os');
const { intro, outro, multiselect, spinner } = require('@clack/prompts');
const pc = require('picocolors');

const SERVER_CONFIG = {
  command: "npx",
  args: ["-y", "aegis-ssh-mcp", "server"]
};

// Paths
const HOMEDIR = os.homedir();
const CLAUDE_MAC = path.join(HOMEDIR, 'Library/Application Support/Claude/claude_desktop_config.json');
const CLAUDE_WIN = path.join(process.env.APPDATA || '', 'Claude', 'claude_desktop_config.json');
const CURSOR_MAC_LINUX = path.join(HOMEDIR, '.cursor', 'mcp.json'); // Note: Cursor handles mcp differently now, but we use the known path

function getClaudePath() {
    if (process.platform === 'darwin') return CLAUDE_MAC;
    if (process.platform === 'win32') return CLAUDE_WIN;
    return null; // Not officially supported on Linux yet
}

function updateJsonConfig(filePath, serverKey, config) {
    let data = {};
    if (fs.existsSync(filePath)) {
        try {
            const content = fs.readFileSync(filePath, 'utf-8');
            data = JSON.parse(content);
        } catch (e) {
            console.error(pc.yellow(`Warning: Could not parse ${filePath}, overwriting.`));
        }
    } else {
        fs.mkdirSync(path.dirname(filePath), { recursive: true });
    }

    if (!data.mcpServers) data.mcpServers = {};
    data.mcpServers[serverKey] = config;

    fs.writeFileSync(filePath, JSON.stringify(data, null, 2), 'utf-8');
}

async function main() {
    intro(pc.inverse(' 🛡️ Aegis SSH Setup '));

    const options = [];
    const claudePath = getClaudePath();
    if (claudePath) {
        options.push({ value: 'claude', label: 'Claude Desktop', hint: 'claude_desktop_config.json' });
    }
    options.push({ value: 'cursor', label: 'Cursor', hint: '~/.cursor/mcp.json' });

    const selected = await multiselect({
        message: 'Which AI clients would you like to install Aegis into?',
        options: options,
        required: false
    });

    if (!selected || selected.length === 0) {
        outro(pc.yellow('Installation cancelled. No clients selected.'));
        process.exit(0);
    }

    const s = spinner();
    s.start('Configuring clients');

    if (selected.includes('claude') && claudePath) {
        updateJsonConfig(claudePath, 'aegis-ssh', SERVER_CONFIG);
    }

    if (selected.includes('cursor')) {
        updateJsonConfig(CURSOR_MAC_LINUX, 'aegis-ssh', SERVER_CONFIG);
    }

    // Generate Skills file
    const rulesPath = path.join(process.cwd(), '.aegis_skills.md');
    fs.writeFileSync(rulesPath, `# Aegis SSH Skills\nYou are a DevOps Engineer. Use the 'aegis-ssh' MCP server tools to:\n- list_servers to view infrastructure.\n- execute_command to run maintenance tasks.\n- run_playbook for idempotent setups.\n- tunnel_expose to proxy local traffic.\n\nKeep interactions minimal and use appropriate shell commands.`, 'utf-8');

    s.stop('Configuration complete');
    
    console.log();
    console.log(pc.green('✔ Successfully installed Aegis SSH!'));
    if (selected.includes('claude')) console.log(`  - Updated Claude config: ${pc.dim(claudePath)}`);
    if (selected.includes('cursor')) console.log(`  - Updated Cursor config: ${pc.dim(CURSOR_MAC_LINUX)}`);
    
    console.log();
    console.log(pc.cyan('Skills injected:'));
    console.log(`  Created ${pc.bold('.aegis_skills.md')} in your current directory.`);
    console.log(`  You can copy its contents into your AI's system prompt or custom instructions.`);
    
    outro("You're all set! Restart your AI client to begin.");
}

main().catch(console.error);
