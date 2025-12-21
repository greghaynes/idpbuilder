#!/bin/bash
set -e

# Build script for generating the static site
# This script prepares the site for deployment to Cloudflare Pages

echo "Building IDP Builder static site..."

# Install npm dependencies if needed
if [ -f "package.json" ]; then
    echo "Checking npm dependencies..."
    if command -v npm >/dev/null 2>&1; then
        # Only install if node_modules doesn't exist or is missing dependencies
        if [ ! -d "node_modules" ] || [ ! -d "node_modules/marked" ]; then
            echo "Installing npm dependencies..."
            npm install --quiet
        fi
    else
        echo "Warning: npm not found. Skipping dependency installation."
    fi
fi

# Create output directory
BUILD_DIR="${BUILD_DIR:-./site}"
OUTPUT_DIR="${OUTPUT_DIR:-./public}"

echo "Source directory: $BUILD_DIR"
echo "Output directory: $OUTPUT_DIR"

# Clean output directory
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Copy all static files
echo "Copying static files..."
cp -r "$BUILD_DIR"/* "$OUTPUT_DIR/"

# Copy organized documentation from docs/ to public/docs/
echo "Copying organized documentation..."
DOCS_SOURCE_DIR="${DOCS_SOURCE_DIR:-./docs}"
if [ -d "$DOCS_SOURCE_DIR" ]; then
    # Create docs directory structure in output
    mkdir -p "$OUTPUT_DIR/docs/specs"
    mkdir -p "$OUTPUT_DIR/docs/implementation"
    mkdir -p "$OUTPUT_DIR/docs/user"
    mkdir -p "$OUTPUT_DIR/docs/images"
    
    # Copy markdown files (will be converted to HTML next)
    cp -r "$DOCS_SOURCE_DIR/specs"/*.md "$OUTPUT_DIR/docs/specs/" 2>/dev/null || true
    cp -r "$DOCS_SOURCE_DIR/implementation"/*.md "$OUTPUT_DIR/docs/implementation/" 2>/dev/null || true
    cp -r "$DOCS_SOURCE_DIR/user"/*.md "$OUTPUT_DIR/docs/user/" 2>/dev/null || true
    cp -r "$DOCS_SOURCE_DIR/images"/* "$OUTPUT_DIR/docs/images/" 2>/dev/null || true
    
    # Copy main docs README if it exists
    [ -f "$DOCS_SOURCE_DIR/README.md" ] && cp "$DOCS_SOURCE_DIR/README.md" "$OUTPUT_DIR/docs/README.md"
    
    echo "Documentation copied successfully!"
    
    # Convert markdown to HTML
    echo "Converting markdown documentation to HTML..."
    if command -v node >/dev/null 2>&1; then
        if [ -f "./scripts/convert-markdown.js" ]; then
            DOCS_SOURCE_DIR="$DOCS_SOURCE_DIR" OUTPUT_DIR="$OUTPUT_DIR" node ./scripts/convert-markdown.js
            
            # Remove the markdown source files after conversion
            find "$OUTPUT_DIR/docs/specs" -name "*.md" -type f ! -name "README.md" -delete 2>/dev/null || true
            find "$OUTPUT_DIR/docs/implementation" -name "*.md" -type f ! -name "README.md" -delete 2>/dev/null || true
            find "$OUTPUT_DIR/docs/user" -name "*.md" -type f ! -name "README.md" -delete 2>/dev/null || true
            
            echo "Markdown conversion completed!"
        else
            echo "Warning: Conversion script not found. Markdown files will be served as-is."
        fi
    else
        echo "Warning: Node.js not found. Markdown files will be served as-is."
    fi
else
    echo "Warning: Documentation source directory not found at $DOCS_SOURCE_DIR"
fi

# Copy examples from examples/ to public/docs/examples/
echo "Copying examples..."
EXAMPLES_SOURCE_DIR="${EXAMPLES_SOURCE_DIR:-./examples}"
if [ -d "$EXAMPLES_SOURCE_DIR" ]; then
    # Create examples directory in output
    mkdir -p "$OUTPUT_DIR/docs/examples"
    mkdir -p "$OUTPUT_DIR/docs/examples/v1alpha2"
    
    # Copy all example files and directories
    cp -r "$EXAMPLES_SOURCE_DIR"/* "$OUTPUT_DIR/docs/examples/" 2>/dev/null || true
    
    echo "Examples copied successfully!"
    
    # Convert example README markdown files to HTML using a custom approach
    if command -v node >/dev/null 2>&1; then
        if [ -d "node_modules/marked" ]; then
            # Create a temporary script to convert examples READMEs
            cat > ./scripts/convert-examples-temp.js << 'EOFJS'
const fs = require('fs');
const path = require('path');
const { marked } = require('marked');

marked.setOptions({ gfm: true, breaks: false, headerIds: true, mangle: false });

const createHtmlPage = (title, content, breadcrumbPath = '') => `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="${title} - IDP Builder Examples">
    <title>${title} | IDP Builder</title>
    <link rel="stylesheet" href="/css/style.css">
    <style>
        .docs-container { max-width: 900px; margin: 2rem auto; padding: 0 1rem; }
        .breadcrumb { font-size: 0.9rem; color: var(--text-muted); margin-bottom: 1rem; }
        .breadcrumb a { color: var(--primary-color); text-decoration: none; }
        .breadcrumb a:hover { text-decoration: underline; }
        .markdown-content { line-height: 1.6; }
        .markdown-content h1 { border-bottom: 2px solid var(--border-color); padding-bottom: 0.5rem; margin-bottom: 1.5rem; }
        .markdown-content h2 { margin-top: 2rem; padding-bottom: 0.3rem; border-bottom: 1px solid var(--border-color); }
        .markdown-content h3 { margin-top: 1.5rem; }
        .markdown-content pre { background-color: #f6f8fa; padding: 1rem; border-radius: 6px; overflow-x: auto; }
        .markdown-content code { font-family: 'Courier New', monospace; background-color: #f6f8fa; padding: 0.2rem 0.4rem; border-radius: 3px; font-size: 0.9em; }
        .markdown-content pre code { background-color: transparent; padding: 0; }
        .markdown-content ul, .markdown-content ol { margin-bottom: 1rem; }
        .markdown-content table { border-collapse: collapse; width: 100%; margin: 1rem 0; }
        .markdown-content table th, .markdown-content table td { border: 1px solid var(--border-color); padding: 0.5rem; }
        .markdown-content table th { background-color: var(--bg-alt); }
        .markdown-content blockquote { border-left: 4px solid var(--primary-color); padding-left: 1rem; margin: 1rem 0; color: var(--text-muted); }
    </style>
</head>
<body>
    <header>
        <nav class="container">
            <div class="logo"><h1>IDP Builder</h1></div>
            <button class="menu-toggle" aria-label="Toggle menu" onclick="toggleMenu()">☰</button>
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
            ${breadcrumbPath ? `<div class="breadcrumb">${breadcrumbPath}</div>` : ''}
            <div class="markdown-content">${content}</div>
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
        document.addEventListener('click', function(event) {
            const nav = document.querySelector('nav');
            const navLinks = document.getElementById('navLinks');
            if (!nav.contains(event.target)) { navLinks.classList.remove('active'); }
        });
    </script>
</body>
</html>`;

const outputDir = process.env.OUTPUT_DIR || './public';

// Convert main examples README
const readmePath = path.join(outputDir, 'docs/examples/README.md');
if (fs.existsSync(readmePath)) {
    const markdown = fs.readFileSync(readmePath, 'utf8');
    const html = marked(markdown);
    const titleMatch = markdown.match(/^#\s+(.+)$/m);
    const title = titleMatch ? titleMatch[1] : 'Examples';
    const breadcrumb = '<a href="/docs">Documentation</a> / <a href="/docs/examples.html">Examples</a> / Platform Examples';
    const fullHtml = createHtmlPage(title, html, breadcrumb);
    const outputPath = path.join(outputDir, 'docs/examples/README.html');
    fs.writeFileSync(outputPath, fullHtml);
    console.log(`Converted: examples/README.md -> ${outputPath}`);
}

// Convert v1alpha2 README
const v1alpha2Path = path.join(outputDir, 'docs/examples/v1alpha2/README.md');
if (fs.existsSync(v1alpha2Path)) {
    const markdown = fs.readFileSync(v1alpha2Path, 'utf8');
    const html = marked(markdown);
    const titleMatch = markdown.match(/^#\s+(.+)$/m);
    const title = titleMatch ? titleMatch[1] : 'V1Alpha2 Examples';
    const breadcrumb = '<a href="/docs">Documentation</a> / <a href="/docs/examples.html">Examples</a> / V1Alpha2 Examples';
    const fullHtml = createHtmlPage(title, html, breadcrumb);
    const outputPath = path.join(outputDir, 'docs/examples/v1alpha2/README.html');
    fs.writeFileSync(outputPath, fullHtml);
    console.log(`Converted: examples/v1alpha2/README.md -> ${outputPath}`);
}

console.log('Examples conversion complete!');
EOFJS
            
            OUTPUT_DIR="$OUTPUT_DIR" node ./scripts/convert-examples-temp.js
            rm -f ./scripts/convert-examples-temp.js
            
            # Remove the markdown source files after conversion
            find "$OUTPUT_DIR/docs/examples" -name "README.md" -type f -delete 2>/dev/null || true
            
            echo "Examples markdown conversion completed!"
        fi
    fi
else
    echo "Warning: Examples source directory not found at $EXAMPLES_SOURCE_DIR"
fi

# Create _headers file for Cloudflare Pages (optional security headers)
echo "Creating _headers file..."
cat > "$OUTPUT_DIR/_headers" << 'EOF'
/*
  X-Frame-Options: DENY
  X-Content-Type-Options: nosniff
  X-XSS-Protection: 1; mode=block
  Referrer-Policy: strict-origin-when-cross-origin
  Permissions-Policy: accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()

/*.css
  Cache-Control: public, max-age=31536000, immutable

/*.js
  Cache-Control: public, max-age=31536000, immutable

/index.html
  Cache-Control: public, max-age=0, must-revalidate
EOF

# Create _redirects file for Cloudflare Pages (optional redirects)
echo "Creating _redirects file..."
cat > "$OUTPUT_DIR/_redirects" << 'EOF'
# Redirect /docs to /docs/index.html
/docs /docs/index.html 200

# Serve markdown files as plain text or allow direct access
/docs/specs/* /docs/specs/:splat 200
/docs/implementation/* /docs/implementation/:splat 200
/docs/user/* /docs/user/:splat 200
/docs/images/* /docs/images/:splat 200
/docs/examples/* /docs/examples/:splat 200
EOF

echo "Build completed successfully!"
echo "Site built to: $OUTPUT_DIR"
