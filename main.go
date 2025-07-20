/*
ObsidianToQuartz - A tool to copy content from Obsidian to Quartz

Features:
- Copies content to a "content" folder in the Quartz directory
- Only copies .svg files from Excalidraw folders
- Transforms Excalidraw links:
  - Wiki-style: [[drawing.excalidraw]] → [[drawing.excalidraw.svg|drawing]]
  - Markdown-style: [text](drawing.excalidraw.md) → [text](drawing.excalidraw.svg)
- Skips all directories starting with . (like .obsidian, .trash)
- Supports exclusion patterns via .obsidian-to-quartz-ignore file

Usage: ObsidianToQuartz <Obsidian_Folder> <Quartz_Folder>
*/

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <Obsidian_Folder> <Quartz_Folder>\n", os.Args[0])
		os.Exit(1)
	}

	obsidianFolder := os.Args[1]
	quartzFolder := os.Args[2]

	// Read exclusion patterns from .obsidian-to-quartz-ignore file
	excludePatterns := readExcludePatterns(obsidianFolder)
	if len(excludePatterns) > 0 {
		fmt.Printf("Loaded %d exclusion patterns\n", len(excludePatterns))
	}

	// Ensure Quartz content folder exists
	contentFolder := filepath.Join(quartzFolder, "content")
	if err := os.MkdirAll(contentFolder, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating content folder: %v\n", err)
		os.Exit(1)
	}

	// Walk through Obsidian folder
	err := filepath.Walk(obsidianFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root folder itself
		if path == obsidianFolder {
			return nil
		}

		// Get relative path from obsidian folder
		relPath, err := filepath.Rel(obsidianFolder, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Skip any directory starting with . (hidden folders like .obsidian, .trash, etc.)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Check if path matches any exclusion pattern
		if shouldExclude(relPath, excludePatterns, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Determine destination path
		destPath := filepath.Join(contentFolder, relPath)

		// Handle directories
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Check if file is in Excalidraw folder and not an SVG
		if isInExcalidrawFolder(relPath) && !strings.HasSuffix(path, ".svg") {
			// Skip non-SVG files in Excalidraw folders
			return nil
		}

		// Process the file
		if strings.HasSuffix(path, ".md") {
			// Process markdown files (transform excalidraw links)
			return processMarkdownFile(path, destPath)
		} else {
			// Copy other files as-is
			return copyFile(path, destPath)
		}
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking through folder: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Conversion completed successfully!")
}

// isInExcalidrawFolder checks if a file path contains "Excalidraw" folder
func isInExcalidrawFolder(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		if strings.EqualFold(part, "Excalidraw") {
			return true
		}
	}
	return false
}

// processMarkdownFile reads a markdown file, transforms excalidraw links, and writes to destination
// Transforms:
//   - [[drawing.excalidraw]] → [[drawing.excalidraw.svg|drawing]]
//   - [text](drawing.excalidraw.md) → [text](drawing.excalidraw.svg)
func processMarkdownFile(src, dest string) error {
	// Read the source file
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %v", err)
	}

	// Replace .excalidraw]] with .excalidraw.svg|name]]
	// This regex captures the filename before .excalidraw
	re := regexp.MustCompile(`\[\[([^|\]]+?)\.excalidraw\]\]`)
	modifiedContent := re.ReplaceAll(content, []byte("[[$1.excalidraw.svg|$1]]"))

	// Also replace markdown-style links: .excalidraw.md) with .excalidraw.svg)
	re2 := regexp.MustCompile(`\.excalidraw\.md\)`)
	modifiedContent = re2.ReplaceAll(modifiedContent, []byte(".excalidraw.svg)"))

	// Ensure destination directory exists
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Write the modified content
	if err := os.WriteFile(dest, modifiedContent, 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %v", err)
	}

	fmt.Printf("Processed: %s -> %s\n", src, dest)
	return nil
}

// copyFile copies a file from src to dest
func copyFile(src, dest string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	// Get file info for permissions
	info, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %v", err)
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Create destination file
	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	// Copy content
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	fmt.Printf("Copied: %s -> %s\n", src, dest)
	return nil
}

// readExcludePatterns reads exclusion patterns from .obsidian-to-quartz-ignore file
func readExcludePatterns(obsidianFolder string) []string {
	ignoreFile := filepath.Join(obsidianFolder, ".obsidian-to-quartz-ignore")
	file, err := os.Open(ignoreFile)
	if err != nil {
		// File doesn't exist, return empty patterns
		return []string{}
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns
}

// shouldExclude checks if a path matches any exclusion pattern
func shouldExclude(relPath string, patterns []string, isDir bool) bool {
	// Normalize path separators for consistent matching
	relPath = filepath.ToSlash(relPath)

	for _, pattern := range patterns {
		pattern = filepath.ToSlash(pattern)

		// Check if pattern is for directories only (ends with /)
		if strings.HasSuffix(pattern, "/") {
			if !isDir {
				continue
			}
			pattern = strings.TrimSuffix(pattern, "/")
		}

		// Check for exact match
		if relPath == pattern {
			return true
		}

		// Check if path starts with pattern (for directory exclusion)
		if isDir && strings.HasPrefix(relPath+"/", pattern+"/") {
			return true
		}

		// Check if any parent directory matches (for file exclusion)
		if !isDir {
			dir := filepath.Dir(relPath)
			if dir != "." && strings.HasPrefix(dir+"/", pattern+"/") {
				return true
			}
		}

		// Simple glob matching for * wildcard
		if strings.Contains(pattern, "*") {
			// Convert simple glob to regex
			regexPattern := strings.ReplaceAll(pattern, "*", ".*")
			regexPattern = "^" + regexPattern + "$"
			if matched, _ := regexp.MatchString(regexPattern, relPath); matched {
				return true
			}
		}
	}

	return false
}
