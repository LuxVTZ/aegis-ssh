document.addEventListener('DOMContentLoaded', () => {
    initRouter();
    fetchData();
    setInterval(fetchData, 5000); // Poll every 5s
});

function initRouter() {
    const links = document.querySelectorAll('.nav-item');
    const sections = document.querySelectorAll('.page-section');

    function navigate() {
        let hash = window.location.hash.substring(1) || 'dashboard';
        
        // Update sidebar active state
        links.forEach(l => {
            if (l.dataset.target === hash) l.classList.add('active');
            else l.classList.remove('active');
        });

        // Update main content visibility
        sections.forEach(s => {
            if (s.id === `section-${hash}`) s.classList.add('active-section');
            else s.classList.remove('active-section');
        });
    }

    window.addEventListener('hashchange', navigate);
    navigate(); // initial load
}

async function fetchData() {
    try {
        const [serversRes, auditRes] = await Promise.all([
            fetch('/api/servers'),
            fetch('/api/audit')
        ]);
        
        const servers = await serversRes.json();
        const audit = await auditRes.json();
        
        updateServersTable(servers);
        updateAuditTable(audit);
        
        document.getElementById('stat-servers').textContent = (servers && Array.isArray(servers)) ? servers.length : 0;
        document.getElementById('stat-cmds').textContent = (audit && Array.isArray(audit)) ? audit.length : 0;
        
    } catch (e) {
        console.error("Failed to fetch dashboard data:", e);
    }
}

function updateServersTable(servers) {
    const tbody = document.querySelector('#servers-table tbody');
    if (!servers || !Array.isArray(servers)) return;
    
    tbody.innerHTML = servers.map(s => `
        <tr>
            <td style="font-weight: 500;">${s.Name}</td>
            <td style="color: var(--text-secondary)">${s.Host}:${s.Port}</td>
            <td style="color: var(--text-secondary)">${s.User}</td>
            <td>${(s.Tags || []).map(t => `<span class="tag">${t}</span>`).join('')}</td>
            <td><span style="color: var(--success); font-size: 12px; font-weight: 500;">Ready</span></td>
        </tr>
    `).join('');
}

function updateAuditTable(logs) {
    const tbody = document.querySelector('#audit-table tbody');
    if (!logs || !Array.isArray(logs)) return;
    
    tbody.innerHTML = logs.map(l => `
        <tr>
            <td style="color: var(--text-secondary); font-variant-numeric: tabular-nums;">
                ${new Date(l.Timestamp).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit', second:'2-digit'})}
            </td>
            <td style="font-weight: 500;">${l.Host}</td>
            <td style="font-family: monospace; color: var(--text-primary)">${l.Command}</td>
            <td class="${l.ExitCode === 0 ? 'exit-success' : 'exit-error'}">
                ${l.ExitCode === 0 ? 'Success' : `Failed (${l.ExitCode})`}
            </td>
            <td style="color: var(--text-secondary); font-variant-numeric: tabular-nums;">${l.DurationMs}ms</td>
        </tr>
    `).join('');
}
