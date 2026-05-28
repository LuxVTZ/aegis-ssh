#!/usr/bin/env node

const fs = require('fs');
const os = require('os');
const path = require('path');
const https = require('https');
const { spawn } = require('child_process');

const VERSION = "v1.0.0"; // Should match GitHub Release tag
const REPO = "LuxVTZ/aegis-ssh";

// Map node platform to Go GOOS
const platformMap = {
    win32: 'windows',
    darwin: 'darwin',
    linux: 'linux'
};

// Map node arch to Go GOARCH
const archMap = {
    x64: 'amd64',
    arm64: 'arm64'
};

const goos = platformMap[process.platform];
const goarch = archMap[process.arch];

if (!goos || !goarch) {
    console.error(`Unsupported platform or architecture: ${process.platform} ${process.arch}`);
    process.exit(1);
}

const isWindows = goos === 'windows';
const binaryName = `aegis-ssh-${goos}-${goarch}${isWindows ? '.exe' : ''}`;
const downloadUrl = `https://github.com/${REPO}/releases/download/${VERSION}/${binaryName}`;

// We store the downloaded binary in OS temp dir so it persists across rapid npx calls
const binDir = path.join(os.tmpdir(), 'aegis-mcp-bin');
const binPath = path.join(binDir, binaryName);

function downloadBinary(url, dest) {
    return new Promise((resolve, reject) => {
        fs.mkdirSync(path.dirname(dest), { recursive: true });
        const file = fs.createWriteStream(dest);

        https.get(url, (response) => {
            if (response.statusCode === 301 || response.statusCode === 302) {
                // Handle redirect
                return downloadBinary(response.headers.location, dest).then(resolve).catch(reject);
            }
            if (response.statusCode !== 200) {
                fs.unlinkSync(dest);
                return reject(new Error(`Failed to download binary: HTTP ${response.statusCode}`));
            }

            response.pipe(file);
            file.on('finish', () => {
                file.close();
                if (!isWindows) {
                    fs.chmodSync(dest, 0o755); // Make executable
                }
                resolve();
            });
        }).on('error', (err) => {
            fs.unlinkSync(dest);
            reject(err);
        });
    });
}

function runBinary() {
    const args = process.argv.slice(2);
    const child = spawn(binPath, args, {
        stdio: 'inherit' // Pass stdin/stdout/stderr directly for MCP communication
    });

    child.on('exit', (code) => {
        process.exit(code);
    });
}

async function main() {
    if (!fs.existsSync(binPath)) {
        // Only log to stderr so we don't pollute MCP stdout which must be pure JSON-RPC
        console.error(`[Aegis NPM] Downloading native binary for ${goos}-${goarch}...`);
        try {
            await downloadBinary(downloadUrl, binPath);
            console.error(`[Aegis NPM] Download complete.`);
        } catch (e) {
            console.error(`[Aegis NPM] Error: ${e.message}`);
            process.exit(1);
        }
    }
    
    runBinary();
}

main();
