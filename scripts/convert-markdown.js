#!/usr/bin/env node

/**
 * Convert markdown documentation to HTML for the static site
 * This script converts all markdown files in docs/ to HTML pages
 */

const fs = require('fs');
const path = require('path');
const { marked } = require('marked');

// Custom renderer for mermaid diagrams and prism.js syntax highlighting
const renderer = new marked.Renderer();
const originalCodeRenderer = renderer.code.bind(renderer);
const originalHeadingRenderer = renderer.heading.bind(renderer);

// Helper function to create slug from text
const slugify = (text) => {
  return text
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_-]+/g, '-')
    .replace(/^-+|-+$/g, '');
};

// Custom heading renderer to ensure IDs are added
renderer.heading = function(text, level, raw) {
  const id = slugify(raw);
  return `<h${level} id="${id}">${text}</h${level}>\n`;
};

// Helper function to sanitize language identifier
const sanitizeLanguage = (lang) => {
  if (!lang) return '';
  // Only allow alphanumeric, dash, and underscore characters
  return lang.replace(/[^a-zA-Z0-9_-]/g, '');
};

// Helper function to escape HTML entities
const escapeHtml = (text) => {
  const map = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#039;'
  };
  return text.replace(/[&<>"']/g, (m) => map[m]);
};

renderer.code = function(code, language) {
  if (language === 'mermaid') {
    // Mermaid diagrams need unescaped code to be rendered by mermaid.js
    // Note: This is safe because mermaid.js will parse and render the content
    return `<pre class="mermaid">${code}</pre>`;
  }
  // Add line-numbers class for Prism.js
  if (language) {
    const safeLang = sanitizeLanguage(language);
    const escapedCode = escapeHtml(code);
    return `<pre class="line-numbers"><code class="language-${safeLang}">${escapedCode}</code></pre>`;
  }
  return originalCodeRenderer(code, language);
};

// Configure marked for GitHub-flavored markdown
marked.setOptions({
  gfm: true,
  breaks: false,
  headerIds: true,
  mangle: false,
  renderer: renderer
});

// Common CSS styles for markdown content
const MARKDOWN_CONTENT_STYLES = `
        .markdown-content {
            line-height: 1.6;
        }
        .markdown-content h1 {
            border-bottom: 2px solid var(--border-color);
            padding-bottom: 0.5rem;
            margin-bottom: 1.5rem;
        }
        .markdown-content h2 {
            margin-top: 2rem;
            margin-bottom: 1rem;
            border-bottom: 1px solid var(--border-color);
            padding-bottom: 0.3rem;
        }
        .markdown-content h3 {
            margin-top: 1.5rem;
            margin-bottom: 0.75rem;
        }
        .markdown-content h4 {
            margin-top: 1.25rem;
            margin-bottom: 0.5rem;
        }
        .markdown-content pre {
            background-color: var(--bg-alt);
            padding: 1rem;
            border-radius: 5px;
            overflow-x: auto;
        }
        .markdown-content code {
            background-color: var(--bg-alt);
            padding: 0.2rem 0.4rem;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
        }
        .markdown-content pre code {
            background-color: transparent;
            padding: 0;
        }
        .markdown-content blockquote {
            border-left: 4px solid var(--primary-color);
            padding-left: 1rem;
            margin-left: 0;
            color: var(--text-muted);
        }
        .markdown-content table {
            border-collapse: collapse;
            width: 100%;
            margin: 1rem 0;
        }
        .markdown-content th,
        .markdown-content td {
            border: 1px solid var(--border-color);
            padding: 0.5rem;
            text-align: left;
        }
        .markdown-content th {
            background-color: var(--bg-alt);
            font-weight: bold;
        }
        .markdown-content a {
            color: var(--primary-color);
            text-decoration: none;
        }
        .markdown-content a:hover {
            text-decoration: underline;
        }
        .markdown-content img {
            max-width: 100%;
            height: auto;
        }
        .markdown-content ul,
        .markdown-content ol {
            padding-left: 2rem;
            margin: 1rem 0;
        }
        .markdown-content li {
            margin: 0.5rem 0;
        }
        /* Mermaid diagram styling */
        .markdown-content .mermaid {
            background-color: transparent;
            padding: 1rem;
            margin: 1.5rem 0;
            text-align: center;
            overflow-x: auto;
        }
        .markdown-content pre.mermaid {
            background-color: var(--bg-alt);
            border-radius: 5px;
        }`;

// Common breadcrumb styles
const BREADCRUMB_STYLES = `
        .breadcrumb {
            font-size: 0.9rem;
            color: var(--text-muted);
            margin-bottom: 1rem;
        }
        .breadcrumb a {
            color: var(--primary-color);
            text-decoration: none;
        }
        .breadcrumb a:hover {
            text-decoration: underline;
        }`;

// Sidebar-specific styles for pages with navigation
const SIDEBAR_STYLES = `
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
        }`;

// Extract headings from markdown for navigation
const extractHeadings = (markdown) => {
  const headings = [];
  const lines = markdown.split('\n');
  
  for (const line of lines) {
    const match = line.match(/^(#{2,4})\s+(.+)$/);
    if (match) {
      const level = match[1].length;
      const text = match[2];
      const id = slugify(text);
      headings.push({ level, text, id });
    }
  }
  
  return headings;
};

// Extract Resource Types (main CRDs) from markdown
const extractResourceTypes = (markdown) => {
  const resourceTypes = { v1alpha1: [], v1alpha2: [] };
  const lines = markdown.split('\n');
  let currentSection = null;
  let inResourceTypes = false;
  
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    
    // Detect which API version section we're in
    if (line.match(/^##\s+idpbuilder\.cnoe\.io\/v1alpha1/)) {
      currentSection = 'v1alpha1';
      inResourceTypes = false;
    } else if (line.match(/^##\s+idpbuilder\.cnoe\.io\/v1alpha2/)) {
      currentSection = 'v1alpha2';
      inResourceTypes = false;
    }
    
    // Detect "Resource Types" section
    if (line.match(/^###\s+Resource Types/)) {
      inResourceTypes = true;
      continue;
    }
    
    // If we hit another h2, h3, or h4, we're out of Resource Types section
    if (inResourceTypes && line.match(/^(##|###|####)\s+/)) {
      inResourceTypes = false;
    }
    
    // Extract resource type links (only from the actual Resource Types list)
    if (inResourceTypes && currentSection && line.match(/^-\s+\[(.+?)\]\(#(.+?)\)/)) {
      const match = line.match(/^-\s+\[(.+?)\]\(#(.+?)\)/);
      const name = match[1];
      const id = match[2];
      resourceTypes[currentSection].push({ name, id });
    }
  }
  
  return resourceTypes;
};

// Build sidebar navigation for API reference
const buildApiSidebar = (resourceTypes) => {
  let html = '';
  
  // Build the sidebar HTML
  html += '<h3>Quick Links</h3>\n';
  html += '<ul>\n';
  html += '  <li><a href="/docs/api.html">API Overview</a></li>\n';
  html += '  <li><a href="#packages">Packages</a></li>\n';
  html += '</ul>\n';
  
  // Add v1alpha1 section if it has items
  if (resourceTypes.v1alpha1.length > 0) {
    html += '<h3>v1alpha1 Resources</h3>\n';
    html += '<ul>\n';
    for (const item of resourceTypes.v1alpha1) {
      html += `  <li><a href="#${item.id}">${item.name}</a></li>\n`;
    }
    html += '</ul>\n';
  }
  
  // Add v1alpha2 section if it has items
  if (resourceTypes.v1alpha2.length > 0) {
    html += '<h3>v1alpha2 Resources</h3>\n';
    html += '<ul>\n';
    for (const item of resourceTypes.v1alpha2) {
      html += `  <li><a href="#${item.id}">${item.name}</a></li>\n`;
    }
    html += '</ul>\n';
  }
  
  return html;
};

// HTML template for documentation pages with sidebar
const createHtmlPageWithSidebar = (title, content, category, sidebar, relativePath = '') => {
  const breadcrumb = category ? `<a href="${relativePath}../index.html">${category}</a>` : '';
  
  return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="${title} - IDP Builder Documentation">
    <title>${title} | IDP Builder</title>
    <link rel="stylesheet" href="${relativePath}../../css/style.css">
    <!-- Prism.js for syntax highlighting -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.css">
    <script type="module">
        import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
        mermaid.initialize({ 
            startOnLoad: true,
            theme: 'default',
            securityLevel: 'loose'
        });
    </script>
    <style>
${SIDEBAR_STYLES}
${BREADCRUMB_STYLES}
${MARKDOWN_CONTENT_STYLES}
    </style>
    <script src="${relativePath}../../js/theme.js"></script>
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
                <li><a href="${relativePath}../../index.html">Home</a></li>
                <li><a href="${relativePath}../index.html">Docs</a></li>
                <li><a href="https://github.com/cnoe-io/idpbuilder" target="_blank" rel="noopener">GitHub</a></li>
            </ul>
        </nav>
    </header>

    <main class="container">
        <div class="docs-container">
            <aside class="docs-sidebar">
${sidebar}
            </aside>

            <article class="docs-content">
                <div class="breadcrumb">
                    <a href="${relativePath}../../index.html">Home</a> / 
                    <a href="${relativePath}../index.html">Docs</a>${breadcrumb ? ' / ' + breadcrumb : ''}
                </div>
                <div class="markdown-content">
                    ${content}
                </div>
            </article>
        </div>
    </main>

    <!-- Theme Toggle Button -->
    <button class="theme-toggle" aria-label="Toggle dark mode">
        <span class="icon light-icon">☀️</span>
        <span class="icon dark-icon">🌙</span>
    </button>

    <footer>
        <div class="container">
            <p>&copy; 2024 CNOE IDP Builder. Licensed under <a href="https://github.com/cnoe-io/idpbuilder/blob/main/LICENSE" target="_blank" rel="noopener">Apache License 2.0</a></p>
        </div>
    </footer>

    <!-- Prism.js for syntax highlighting -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-core.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/autoloader/prism-autoloader.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.js"></script>
    <script>
        function toggleMenu() {
            const navLinks = document.getElementById('navLinks');
            navLinks.classList.toggle('active');
        }

        document.addEventListener('click', function(event) {
            const nav = document.querySelector('nav');
            const navLinks = document.getElementById('navLinks');
            
            if (!nav.contains(event.target)) {
                navLinks.classList.remove('active');
            }
        });
    </script>
</body>
</html>`;
};

// HTML template for documentation pages (without sidebar)
const createHtmlPage = (title, content, category, relativePath = '') => {
  const breadcrumb = category ? `<a href="${relativePath}../index.html">${category}</a>` : '';
  
  return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="${title} - IDP Builder Documentation">
    <title>${title} | IDP Builder</title>
    <link rel="stylesheet" href="${relativePath}../../css/style.css">
    <!-- Prism.js for syntax highlighting -->
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.css">
    <script type="module">
        import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
        mermaid.initialize({ 
            startOnLoad: true,
            theme: 'default',
            securityLevel: 'loose'
        });
    </script>
    <style>
        .docs-container {
            max-width: 900px;
            margin: 2rem auto;
            padding: 0 1rem;
        }
${BREADCRUMB_STYLES}
${MARKDOWN_CONTENT_STYLES}
    </style>
    <script src="${relativePath}../../js/theme.js"></script>
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
                <li><a href="${relativePath}../../index.html">Home</a></li>
                <li><a href="${relativePath}../index.html">Docs</a></li>
                <li><a href="https://github.com/cnoe-io/idpbuilder" target="_blank" rel="noopener">GitHub</a></li>
            </ul>
        </nav>
    </header>

    <main class="container">
        <div class="docs-container">
            <div class="breadcrumb">
                <a href="${relativePath}../../index.html">Home</a> / 
                <a href="${relativePath}../index.html">Docs</a>${breadcrumb ? ' / ' + breadcrumb : ''}
            </div>
            <article class="markdown-content">
                ${content}
            </article>
        </div>
    </main>

    <!-- Theme Toggle Button -->
    <button class="theme-toggle" aria-label="Toggle dark mode">
        <span class="icon light-icon">☀️</span>
        <span class="icon dark-icon">🌙</span>
    </button>

    <footer>
        <div class="container">
            <p>&copy; 2024 CNOE IDP Builder. Licensed under <a href="https://github.com/cnoe-io/idpbuilder/blob/main/LICENSE" target="_blank" rel="noopener">Apache License 2.0</a></p>
        </div>
    </footer>

    <!-- Prism.js for syntax highlighting -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-core.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/autoloader/prism-autoloader.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.js"></script>
    <script>
        function toggleMenu() {
            const navLinks = document.getElementById('navLinks');
            navLinks.classList.toggle('active');
        }

        document.addEventListener('click', function(event) {
            const nav = document.querySelector('nav');
            const navLinks = document.getElementById('navLinks');
            
            if (!nav.contains(event.target)) {
                navLinks.classList.remove('active');
            }
        });
    </script>
</body>
</html>`;
};

// Convert a single markdown file to HTML
const convertMarkdownFile = (inputPath, outputPath, category) => {
  try {
    const markdown = fs.readFileSync(inputPath, 'utf8');
    const html = marked(markdown);
    
    // Extract title from first h1 or use filename
    const titleMatch = markdown.match(/^#\s+(.+)$/m);
    const title = titleMatch ? titleMatch[1] : path.basename(inputPath, '.md');
    
    // Calculate relative path based on nesting
    const relativePath = '';
    
    // Check if this is the API reference page - if so, add sidebar
    const isApiReference = path.basename(inputPath) === 'reference.md' && inputPath.includes('/api/');
    
    let fullHtml;
    if (isApiReference) {
      // Extract resource types and build sidebar for API reference
      const resourceTypes = extractResourceTypes(markdown);
      const sidebar = buildApiSidebar(resourceTypes);
      fullHtml = createHtmlPageWithSidebar(title, html, category, sidebar, relativePath);
    } else {
      fullHtml = createHtmlPage(title, html, category, relativePath);
    }
    
    // Create output directory if it doesn't exist
    const outputDir = path.dirname(outputPath);
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }
    
    fs.writeFileSync(outputPath, fullHtml);
    console.log(`Converted: ${inputPath} -> ${outputPath}`);
  } catch (error) {
    console.error(`Error converting ${inputPath}:`, error.message);
  }
};

// Convert all markdown files in a directory
const convertDirectory = (inputDir, outputDir, category) => {
  if (!fs.existsSync(inputDir)) {
    console.log(`Directory not found: ${inputDir}`);
    return;
  }
  
  const files = fs.readdirSync(inputDir);
  
  files.forEach(file => {
    const inputPath = path.join(inputDir, file);
    const stat = fs.statSync(inputPath);
    
    if (stat.isDirectory()) {
      // Skip subdirectories for now
      return;
    }
    
    if (file.endsWith('.md') && file !== 'README.md') {
      const outputPath = path.join(outputDir, file.replace('.md', '.html'));
      convertMarkdownFile(inputPath, outputPath, category);
    }
  });
};

// Main conversion process
const main = () => {
  const docsSource = process.env.DOCS_SOURCE_DIR || './docs';
  const outputBase = process.env.OUTPUT_DIR || './public';
  const outputDocs = path.join(outputBase, 'docs');
  
  console.log('Converting markdown documentation to HTML...');
  console.log(`Source: ${docsSource}`);
  console.log(`Output: ${outputDocs}`);
  
  // Convert each category
  const categories = [
    { dir: 'specs', title: 'Technical Specifications' },
    { dir: 'implementation', title: 'Implementation Documentation' },
    { dir: 'user', title: 'User Guides' },
    { dir: 'api', title: 'API Reference' }
  ];
  
  categories.forEach(({ dir, title }) => {
    const inputDir = path.join(docsSource, dir);
    const outputDir = path.join(outputDocs, dir);
    convertDirectory(inputDir, outputDir, title);
  });
  
  console.log('Conversion complete!');
};

// Run if called directly
if (require.main === module) {
  main();
}

module.exports = { convertMarkdownFile, convertDirectory };
