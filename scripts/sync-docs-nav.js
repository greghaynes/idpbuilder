#!/usr/bin/env node
/**
 * Automatically sync documentation navigation in site/docs/index.html
 * with the actual markdown files in docs/ directories.
 * 
 * This script:
 * 1. Scans docs/specs, docs/implementation, and docs/user for .md files
 * 2. Extracts title from each markdown file (first # heading)
 * 3. Updates the navigation sidebar in site/docs/index.html
 * 
 * Usage:
 *   node scripts/sync-docs-nav.js
 */

const fs = require('fs');
const path = require('path');

// Paths
const REPO_ROOT = path.join(__dirname, '..');
const DOCS_DIR = path.join(REPO_ROOT, 'docs');
const SITE_INDEX = path.join(REPO_ROOT, 'site', 'docs', 'index.html');

/**
 * Extract the first H1 heading from a markdown file
 * @param {string} filePath - Path to markdown file
 * @returns {string} - Title or filename if no title found
 */
function extractTitle(filePath) {
    try {
        const content = fs.readFileSync(filePath, 'utf-8');
        const titleMatch = content.match(/^#\s+(.+)$/m);
        if (titleMatch) {
            return titleMatch[1].trim();
        }
    } catch (err) {
        console.warn(`Warning: Could not read ${filePath}:`, err.message);
    }
    
    // Fallback to filename
    const filename = path.basename(filePath, '.md');
    // Convert kebab-case to Title Case
    return filename
        .split('-')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
}

/**
 * Get all markdown files in a directory
 * @param {string} category - Directory name (specs, implementation, user)
 * @returns {Array} - Array of {filename, title} objects
 */
function getDocsInCategory(category) {
    const categoryDir = path.join(DOCS_DIR, category);
    
    if (!fs.existsSync(categoryDir)) {
        console.warn(`Warning: Directory not found: ${categoryDir}`);
        return [];
    }
    
    const files = fs.readdirSync(categoryDir)
        .filter(file => file.endsWith('.md') && file !== 'README.md')
        .sort();
    
    return files.map(filename => {
        const filePath = path.join(categoryDir, filename);
        const title = extractTitle(filePath);
        const htmlFile = filename.replace('.md', '.html');
        
        return {
            filename,
            htmlFile,
            title,
            href: `/docs/${category}/${htmlFile}`
        };
    });
}

/**
 * Generate navigation HTML for a category
 * @param {Array} docs - Array of doc objects from getDocsInCategory
 * @returns {string} - HTML list items
 */
function generateNavItems(docs) {
    if (docs.length === 0) {
        return '                        <li><em>No documents available</em></li>';
    }
    
    return docs.map(doc => 
        `                        <li><a href="${doc.href}">${doc.title}</a></li>`
    ).join('\n');
}

/**
 * Update the navigation in site/docs/index.html
 */
function updateNavigation() {
    console.log('Syncing documentation navigation...\n');
    
    // Get docs from each category
    const specs = getDocsInCategory('specs');
    const implementation = getDocsInCategory('implementation');
    const user = getDocsInCategory('user');
    
    console.log(`Found ${specs.length} spec(s)`);
    console.log(`Found ${implementation.length} implementation doc(s)`);
    console.log(`Found ${user.length} user guide(s)\n`);
    
    // Read the current index.html
    if (!fs.existsSync(SITE_INDEX)) {
        console.error(`Error: Navigation file not found: ${SITE_INDEX}`);
        process.exit(1);
    }
    
    let html = fs.readFileSync(SITE_INDEX, 'utf-8');
    
    // Replace Technical Specs section
    const specsNav = generateNavItems(specs);
    html = html.replace(
        /(<details>\s*<summary>Technical Specs<\/summary>\s*<ul>\s*\n)([\s\S]*?)(\s*<\/ul>\s*<\/details>)/m,
        `$1${specsNav}\n$3`
    );
    
    // Replace Implementation Docs section
    const implNav = generateNavItems(implementation);
    html = html.replace(
        /(<details>\s*<summary>Implementation Docs<\/summary>\s*<ul>\s*\n)([\s\S]*?)(\s*<\/ul>\s*<\/details>)/m,
        `$1${implNav}\n$3`
    );
    
    // Replace User Guides section
    const userNav = generateNavItems(user);
    html = html.replace(
        /(<details( open)?>\s*<summary>User Guides<\/summary>\s*<ul>\s*\n)([\s\S]*?)(\s*<\/ul>\s*<\/details>)/m,
        `$1${userNav}\n$4`
    );
    
    // Write back
    fs.writeFileSync(SITE_INDEX, html, 'utf-8');
    
    console.log('✅ Navigation updated successfully!');
    console.log(`Updated: ${SITE_INDEX}\n`);
    
    // Print summary
    console.log('Navigation now includes:');
    console.log(`  - ${specs.length} Technical Spec(s)`);
    specs.forEach(doc => console.log(`    • ${doc.title}`));
    console.log(`  - ${implementation.length} Implementation Doc(s)`);
    implementation.forEach(doc => console.log(`    • ${doc.title}`));
    console.log(`  - ${user.length} User Guide(s)`);
    user.forEach(doc => console.log(`    • ${doc.title}`));
}

// Main execution
try {
    updateNavigation();
} catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
}
