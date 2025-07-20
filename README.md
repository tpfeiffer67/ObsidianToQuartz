# ObsidianToQuartz

A Go utility to convert and copy content from an Obsidian vault to a Quartz website folder with specific transformations for Excalidraw drawings.

## Settings used

### Excalidraw plugin  
Excalidraw plugin settings:
- Excalidraw
   - Options
      - Embedding Excalidraw into yout Notes and Exporting
         - Export Settings
            - Auto-Export Settings
               - Auto-export SVG: **Enabled**
### Line breaks
#### Obsidian Options
   - Editor
      - Strict line breaks: **Turned off**

#### Quartz plugin

Edit file `quartz.config.ts` to add plugin HardLineBreaks in section transformers:
```json
  plugins: {
    transformers: [
      Plugin.HardLineBreaks(), 
	  ...
```

## Features

- **Selective Copying**: Copies content from Obsidian to Quartz's `content` folder
- **Excalidraw Handling**: Only copies `.svg` files from Excalidraw folders (ignores other Excalidraw files)
- **Link Transformation**: Automatically transforms Excalidraw links in markdown files:
  - Wiki-style: `[[drawing.excalidraw]]` → `[[drawing.excalidraw.svg|drawing]]`
  - Markdown-style: `[text](drawing.excalidraw.md)` → `[text](drawing.excalidraw.svg)`
  - The wiki links display only the drawing name, while markdown links preserve the original text
- **Hidden Folders**: Skips all directories starting with `.` (like `.obsidian`, `.trash`, etc.)
- **Custom Exclusions**: Support for `.obsidian-to-quartz-ignore` file to exclude specific folders and files
- **Structure Preservation**: Maintains the original folder structure in the destination

## Installation

### Prerequisites
- Go 1.16 or higher installed on your system

### Install

```
go install github.com/tpfeiffer67/ObsidianToQuartz
```

### Building from Source

1. Clone or download the source code
2. Navigate to the project directory
3. Build the executable:

```bash
go build -o ObsidianToQuartz main.go
```

On Windows:
```bash
go build -o ObsidianToQuartz.exe main.go
```

## Usage

```bash
ObsidianToQuartz <Obsidian_Folder> <Quartz_Folder>
```

### Examples

```bash
# Convert an Obsidian vault to Quartz
./ObsidianToQuartz ~/Documents/MyVault ~/Sites/MyQuartzSite

# On Windows
ObsidianToQuartz.exe C:\Users\Me\MyVault C:\Sites\MyQuartzSite
```

The tool will:
1. Create a `content` folder inside your Quartz folder (if it doesn't exist)
2. Check for exclusion patterns in `.obsidian-to-quartz-ignore` file
3. Copy all relevant files while applying the transformation rules
4. Display progress for each file processed

## Excluding Files and Folders

You can exclude specific files and folders by creating a `.obsidian-to-quartz-ignore` file in your Obsidian vault root. This file works similarly to `.gitignore`.

### Syntax

Create a file named `.obsidian-to-quartz-ignore` in your Obsidian vault root:

```
# Comments start with #
# Exclude specific folders
Templates/
Archive/
Private Notes/

# Exclude all draft folders anywhere
*/drafts/

# Exclude specific files
todo.md
personal-journal.md

# Use wildcards
*.tmp
*.backup
*-draft.md
```

### Exclusion Rules

- Lines starting with `#` are treated as comments
- Empty lines are ignored
- Patterns ending with `/` match only directories
- Patterns without `/` match both files and directories
- Simple wildcards (`*`) are supported
- Paths are relative to the Obsidian vault root

### Example `.obsidian-to-quartz-ignore`

```
# Development and testing
Test Vault/
Sandbox/

# Personal content
Personal/
Journal/
Private/

# Templates and resources
Templates/
_templates/

# Temporary files
*.tmp
*.bak
~*

# Draft content
*-draft.md
Drafts/
```

## How It Works

### File Processing Rules

1. **Exclusion Patterns**:
   - Checks `.obsidian-to-quartz-ignore` file first
   - Skips any files or folders matching the patterns

2. **Markdown Files (`.md`)**:
   - Copied with link transformations
   - Wiki-style Excalidraw links are converted to SVG with clean display names
   - Markdown-style Excalidraw links are converted to point to SVG files

3. **Excalidraw Folders**:
   - Only `.svg` files are copied
   - All other files (`.excalidraw`, `.png`, etc.) are ignored

4. **Hidden Directories**:
   - Any folder starting with `.` is completely skipped
   - This includes `.obsidian`, `.trash`, and any other hidden folders

5. **Other Files**:
   - All other files are copied as-is, preserving the directory structure

### Link Transformation Example

If your Obsidian note contains:
```markdown
Check out this diagram: [[Architecture.excalidraw]]
See also [this flowchart](Workflow.excalidraw.md)
```

It will be transformed to:
```markdown
Check out this diagram: [[Architecture.excalidraw.svg|Architecture]]
See also [this flowchart](Workflow.excalidraw.svg)
```

This ensures:
- Wiki-style links work in Quartz with clean display names (just "Architecture")
- Markdown-style links point to the correct `.svg` files
- All links properly reference the SVG files that will be copied

## Output

The tool provides console output showing:
- Number of exclusion patterns loaded (if any)
- Each file being processed or copied
- Any errors encountered
- Success message upon completion

Example output:
```
Loaded 4 exclusion patterns
Copied: /path/to/obsidian/note.md -> /path/to/quartz/content/note.md
Processed: /path/to/obsidian/ideas.md -> /path/to/quartz/content/ideas.md
Copied: /path/to/obsidian/Excalidraw/diagram.svg -> /path/to/quartz/content/Excalidraw/diagram.svg
Conversion completed successfully!
```

## Error Handling

The tool will exit with an error message if:
- Wrong number of arguments provided
- Source folder doesn't exist
- Insufficient permissions to read/write files
- Any file operation fails

## License

MIT License - see LICENSE file for details.