#!/usr/bin/env node

/**
 * Generate API overview page from CRD YAML files
 */

const fs = require('fs');
const path = require('path');

// Simple YAML parser for CRD files
function parseCRDYaml(content) {
    const crds = [];
    const docs = content.split(/^---$/m).filter(d => d.trim());
    
    for (const doc of docs) {
        const lines = doc.split('\n');
        let crd = { metadata: {}, spec: { names: {}, versions: [] } };
        let inSpec = false;
        let inVersions = false;
        let currentVersion = null;
        
        for (const line of lines) {
            if (line.match(/^kind:\s*CustomResourceDefinition/)) {
                crd.kind = 'CustomResourceDefinition';
            } else if (line.match(/^spec:/)) {
                inSpec = true;
            } else if (inSpec && line.match(/^\s{2}group:\s*(.+)/)) {
                crd.spec.group = line.match(/group:\s*(.+)/)[1].trim();
            } else if (inSpec && line.match(/^\s{4}kind:\s*(.+)/)) {
                crd.spec.names.kind = line.match(/kind:\s*(.+)/)[1].trim();
            } else if (inSpec && line.match(/^\s{4}plural:\s*(.+)/)) {
                crd.spec.names.plural = line.match(/plural:\s*(.+)/)[1].trim();
            } else if (inSpec && line.match(/^\s{2}versions:/)) {
                inVersions = true;
            } else if (inVersions && line.match(/^\s{2,4}- name:\s*(.+)/)) {
                if (currentVersion) {
                    crd.spec.versions.push(currentVersion);
                }
                currentVersion = { name: line.match(/name:\s*(.+)/)[1].trim() };
            } else if (inVersions && line.match(/^\s{4}name:\s*(.+)/) && !currentVersion) {
                // Handle non-list format
                currentVersion = { name: line.match(/name:\s*(.+)/)[1].trim() };
            } else if (currentVersion && line.match(/^\s{8,12}description:\s*(.+)/)) {
                currentVersion.description = line.match(/description:\s*(.+)/)[1].trim();
            }
        }
        
        if (currentVersion) {
            crd.spec.versions.push(currentVersion);
        }
        
        if (crd.kind === 'CustomResourceDefinition' && crd.spec.names.kind) {
            crds.push(crd);
        }
    }
    
    return crds;
}

// Read and parse all CRD YAML files
const crdDir = path.join(__dirname, '../pkg/controllers/resources');
const crdFiles = fs.readdirSync(crdDir).filter(f => f.endsWith('.yaml') && f.startsWith('idpbuilder.cnoe.io_'));

const allCrds = [];
for (const file of crdFiles) {
    const content = fs.readFileSync(path.join(crdDir, file), 'utf8');
    const crds = parseCRDYaml(content);
    allCrds.push(...crds);
}

// Extract CRD info
const crdInfo = allCrds.map(crd => {
    const version = crd.spec.versions[0] || {};
    return {
        kind: crd.spec.names.kind,
        group: crd.spec.group,
        apiVersion: version.name,
        description: version.description || `${crd.spec.names.kind} is the Schema for the ${crd.spec.names.plural} API`,
        plural: crd.spec.names.plural,
        anchor: crd.spec.names.kind.toLowerCase()
    };
});

// Group CRDs by API version
const v1alpha1 = crdInfo.filter(c => c.apiVersion === 'v1alpha1');
const v1alpha2 = crdInfo.filter(c => c.apiVersion === 'v1alpha2');

// Generate CRD list HTML for a version
function generateCRDList(crds) {
    return crds.map(crd => 
        `                        <li><code>${crd.kind}</code> - ${crd.description}</li>`
    ).join('\n');
}

// Generate CRD cards for detailed info
function generateCRDCards(crds) {
    return crds.map(crd => `
                <div class="api-card">
                    <h4>${crd.kind}</h4>
                    <p>${crd.description}</p>
                    <p><a href="/docs/api/reference.html#${crd.anchor}">View ${crd.kind} reference →</a></p>
                </div>`).join('\n');
}

// Generate HTML
const apiPageContent = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="IDP Builder API Reference - Custom Resource Definitions (CRDs)">
    <title>API Reference | IDP Builder</title>
    <link rel="stylesheet" href="/css/style.css">
    <style>
        .docs-container {
            display: grid;
            grid-template-columns: 250px 1fr;
            gap: 2rem;
            margin-top: 2rem;
        }
        .docs-sidebar {
            position: sticky;
            top: 80px;
            height: fit-content;
            max-height: calc(100vh - 80px);
            overflow-y: auto;
        }
        .docs-sidebar h3 {
            font-size: 0.9rem;
            text-transform: uppercase;
            color: var(--text-muted);
            margin-top: 1.5rem;
            margin-bottom: 0.5rem;
            padding-left: 0.5rem;
        }
        .docs-sidebar h3:first-child {
            margin-top: 0;
        }
        .docs-sidebar ul {
            list-style: none;
            padding: 0;
        }
        .docs-sidebar li {
            margin-bottom: 0.5rem;
        }
        .docs-sidebar a {
            text-decoration: none;
            color: var(--text-color);
            padding: 0.5rem;
            display: block;
            border-radius: 4px;
            transition: background-color 0.3s;
            font-size: 0.95rem;
        }
        .docs-sidebar a:hover,
        .docs-sidebar a.active {
            background-color: var(--bg-alt);
            color: var(--primary-color);
        }
        .docs-content {
            max-width: 800px;
            width: 100%;
        }
        .docs-content h1 {
            margin-bottom: 1rem;
        }
        .docs-content h2 {
            margin-top: 2rem;
            margin-bottom: 1rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid var(--border-color);
        }
        .docs-content p {
            margin-bottom: 1rem;
        }
        .docs-content ul {
            margin-bottom: 1rem;
            padding-left: 2rem;
        }
        .api-card {
            background-color: var(--bg-alt);
            padding: 1.5rem;
            border-radius: 5px;
            margin-bottom: 1.5rem;
        }
        .api-card h3, .api-card h4 {
            margin-top: 0;
            margin-bottom: 0.5rem;
        }
        .api-card p {
            margin-bottom: 0.5rem;
        }
        .api-card code {
            background-color: var(--bg-color);
            padding: 0.2rem 0.4rem;
            border-radius: 3px;
            font-size: 0.9em;
        }
        @media (max-width: 768px) {
            .docs-container {
                grid-template-columns: 1fr;
            }
            .docs-sidebar {
                position: relative;
                top: 0;
                background-color: var(--bg-alt);
                padding: 1rem;
                border-radius: 5px;
                margin-bottom: 1rem;
            }
            .docs-sidebar h3 {
                margin-top: 0;
            }
            .docs-content {
                max-width: 100%;
                overflow-x: hidden;
            }
        }
    </style>
</head>
<body>
    <header>
        <nav class="container">
            <div class="logo">
                <h1>IDP Builder</h1>
            </div>
            <button class="menu-toggle" aria-label="Toggle menu" onclick="toggleMenu()">
                ☰
            </button>
            <ul class="nav-links" id="navLinks">
                <li><a href="/">Home</a></li>
                <li><a href="/docs">Docs</a></li>
                <li><a href="/docs/examples.html">Examples</a></li>
                <li><a href="https://github.com/cnoe-io/idpbuilder" target="_blank" rel="noopener">GitHub</a></li>
            </ul>
        </nav>
    </header>

    <main class="container">
        <div class="docs-container">
            <aside class="docs-sidebar">
                <h3>API Reference</h3>
                <ul>
                    <li><a href="/docs/api.html" class="active">Overview</a></li>
                    <li><a href="/docs/api/reference.html">CRD Reference</a></li>
                </ul>
                
                <h3>Documentation</h3>
                <ul>
                    <li><a href="/docs">Docs Home</a></li>
                    <li><a href="/docs/examples.html">Examples</a></li>
                </ul>
            </aside>

            <article class="docs-content">
                <h1>API Reference</h1>
                <p>IDP Builder provides ${crdInfo.length} Custom Resource Definitions (CRDs) that allow you to declaratively manage your internal developer platform on Kubernetes.</p>
                
                <div style="background-color: var(--bg-alt); padding: 1rem; border-radius: 5px; margin-bottom: 2rem;">
                    <h3 style="margin-top: 0;">📖 About This Documentation</h3>
                    <p style="margin-bottom: 0;">This API reference is automatically generated from the CRD definitions. It provides detailed information about all fields, types, and validation rules for each Custom Resource.</p>
                </div>

                <h2>API Versions</h2>
                <p>IDP Builder provides two API versions:</p>
                
                <div class="api-card">
                    <h3>v1alpha1</h3>
                    <p><strong>Status:</strong> Deprecated (transitioning to v1alpha2)</p>
                    <p>The original API version with monolithic resources. Includes ${v1alpha1.length} resource(s):</p>
                    <ul>
${generateCRDList(v1alpha1)}
                    </ul>
                </div>

                <div class="api-card">
                    <h3>v1alpha2</h3>
                    <p><strong>Status:</strong> Current (recommended)</p>
                    <p>The new modular architecture with pluggable providers. Includes ${v1alpha2.length} resource(s):</p>
                    <ul>
${generateCRDList(v1alpha2)}
                    </ul>
                </div>

                <h2>Custom Resources</h2>
                <p>The following Custom Resources are available:</p>
${generateCRDCards(crdInfo)}

                <h2>Using the API</h2>
                <p>You can create these resources in two ways:</p>
                
                <h3>1. Using the CLI (Development Mode)</h3>
                <p>The CLI automatically creates appropriate resources based on your requirements:</p>
                <pre><code>idpbuilder create</code></pre>
                
                <h3>2. Declaratively via kubectl (Production Mode)</h3>
                <p>Create resources directly in your Kubernetes cluster:</p>
                <pre><code>kubectl apply -f platform.yaml</code></pre>

                <h2>Examples</h2>
                <p>For example YAML configurations, see the <a href="/docs/examples.html">Examples section</a>:</p>
                <ul>
                    <li><a href="/docs/examples/platform-simple.html">Simple Platform</a></li>
                    <li><a href="/docs/examples/v1alpha2/platform-with-gateway.html">Platform with Gateway</a></li>
                    <li><a href="/docs/examples/v1alpha2/giteaprovider.html">GiteaProvider</a></li>
                </ul>

                <h2>Complete Reference</h2>
                <p>For complete field-level documentation, see the <a href="/docs/api/reference.html">Full CRD Reference Documentation</a>.</p>
            </article>
        </div>
    </main>

    <footer>
        <div class="container">
            <p>&copy; 2024 CNOE IDP Builder. Licensed under <a href="https://github.com/cnoe-io/idpbuilder/blob/main/LICENSE" target="_blank" rel="noopener">Apache License 2.0</a></p>
        </div>
    </footer>

    <script>
        function toggleMenu() {
            const navLinks = document.getElementById('navLinks');
            navLinks.classList.toggle('active');
        }

        // Close menu when clicking outside
        document.addEventListener('click', function(event) {
            const nav = document.querySelector('nav');
            const navLinks = document.getElementById('navLinks');
            
            if (!nav.contains(event.target)) {
                navLinks.classList.remove('active');
            }
        });
    </script>
</body>
</html>
`;

const outputDir = path.join(__dirname, '../site/docs');
const outputPath = path.join(outputDir, 'api.html');

// Ensure output directory exists
if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
}

// Write the file
fs.writeFileSync(outputPath, apiPageContent);
console.log(`Generated: ${outputPath}`);
console.log(`Found ${crdInfo.length} CRDs (${v1alpha1.length} v1alpha1, ${v1alpha2.length} v1alpha2)`);
