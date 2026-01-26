package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackchuka/gh-md/internal/config"
	"github.com/jackchuka/gh-md/internal/github"
	"gopkg.in/yaml.v3"
)

// ParsedComment represents a parsed comment from the markdown file.
type ParsedComment struct {
	ID       string // empty = new comment
	Author   string
	Body     string
	ParentID string // for discussion replies (derived from indentation)
}

// ParsedFile represents a parsed markdown file.
type ParsedFile struct {
	// From frontmatter
	ID      string
	Owner   string
	Repo    string
	Number  int
	Updated time.Time // For conflict detection
	State   string    // open/closed from frontmatter

	// From content
	Title    string
	Body     string
	ItemType github.ItemType
	Comments []ParsedComment // Parsed from comments section

	// Original file path
	FilePath string
}

// frontmatter represents the YAML frontmatter structure.
type frontmatter struct {
	ID      string    `yaml:"id"`
	Owner   string    `yaml:"owner"`
	Repo    string    `yaml:"repo"`
	Number  int       `yaml:"number"`
	Updated time.Time `yaml:"updated"`
	State   string    `yaml:"state"`
}

// ParseFile parses a markdown file and returns structured data.
func ParseFile(path string) (*ParsedFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return parseContent(string(content), path)
}

func parseContent(content, path string) (*ParsedFile, error) {
	// Extract frontmatter
	fm, rest, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Extract title and body
	title, body := extractTitleAndBody(rest)

	// Extract comments
	comments := parseComments(rest)

	// Detect item type from path
	itemType := detectItemType(path)

	return &ParsedFile{
		ID:       fm.ID,
		Owner:    fm.Owner,
		Repo:     fm.Repo,
		Number:   fm.Number,
		Updated:  fm.Updated,
		State:    fm.State,
		Title:    title,
		Body:     body,
		ItemType: itemType,
		Comments: comments,
		FilePath: path,
	}, nil
}

func extractFrontmatter(content string) (*frontmatter, string, error) {
	// Frontmatter is between --- markers
	if !strings.HasPrefix(content, "---\n") {
		return nil, "", fmt.Errorf("file does not start with frontmatter")
	}

	// Find the closing ---
	endIndex := strings.Index(content[4:], "\n---\n")
	if endIndex == -1 {
		return nil, "", fmt.Errorf("frontmatter not closed")
	}

	fmContent := content[4 : 4+endIndex]
	rest := content[4+endIndex+5:] // Skip past closing ---\n

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(fmContent), &fm); err != nil {
		return nil, "", fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return &fm, rest, nil
}

const (
	contentStart = "<!-- gh-md:content -->"
	contentEnd   = "<!-- /gh-md:content -->"
)

func extractTitleAndBody(content string) (string, string) {
	startIdx := strings.Index(content, contentStart)
	endIdx := strings.Index(content, contentEnd)

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return "", ""
	}

	editableContent := strings.TrimSpace(content[startIdx+len(contentStart) : endIdx])

	// Title is the first # heading
	if !strings.HasPrefix(editableContent, "# ") {
		return "", editableContent
	}

	newlineIdx := strings.Index(editableContent, "\n")
	if newlineIdx == -1 {
		return strings.TrimPrefix(editableContent, "# "), ""
	}

	title := strings.TrimPrefix(editableContent[:newlineIdx], "# ")
	body := strings.TrimSpace(editableContent[newlineIdx+1:])
	return title, body
}

func detectItemType(path string) github.ItemType {
	dir := filepath.Dir(path)
	base := filepath.Base(dir)
	if itemType, ok := github.ItemTypeFromDirName(base); ok {
		return itemType
	}
	return ""
}

// ResolveFilePath resolves a URL, short path, or file path to an actual file path.
// Supports:
//   - Full URL: https://github.com/owner/repo/issues/123
//   - Short path: owner/repo/issues/123
//   - Root-relative path: owner/repo/issues/123.md
//   - Local file: ~/.gh-md/owner/repo/issues/123.md
func ResolveFilePath(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty input")
	}

	// It's a file path - check if it exists as-is first
	if _, err := os.Stat(input); err == nil {
		return input, nil
	}

	// Try interpreting it as a path relative to GH_MD_ROOT (~/.gh-md by default).
	if path, err := resolveRootRelativePath(input); err == nil {
		return path, nil
	}

	parsed, err := github.ParseInput(input)
	if err == nil && parsed.Number > 0 && parsed.ItemType != "" {
		itemDir, ok := parsed.ItemType.DirName()
		if !ok {
			return "", fmt.Errorf("unsupported item type: %s", parsed.ItemType)
		}

		root, err := config.GetRootDir()
		if err != nil {
			return "", err
		}

		expected := filepath.Join(root, parsed.Owner, parsed.Repo, itemDir, fmt.Sprintf("%d.md", parsed.Number))
		if _, err := os.Stat(expected); err == nil {
			return expected, nil
		}
		return "", fmt.Errorf("local file not found: %s (run 'gh md pull' first)", expected)
	}

	return "", fmt.Errorf("file not found: %s", input)
}

func resolveRootRelativePath(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty path")
	}

	// Only try this for relative paths to avoid surprises.
	if filepath.IsAbs(input) {
		return "", fmt.Errorf("absolute path")
	}

	root, err := config.GetRootDir()
	if err != nil {
		return "", err
	}

	candidates := []string{
		filepath.Join(root, input),
	}
	if !strings.HasSuffix(input, ".md") {
		candidates = append(candidates, filepath.Join(root, input+".md"))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("local file not found under root: %s", root)
}

const (
	commentStart    = "<!-- gh-md:comment\n" // newline distinguishes from <!-- gh-md:comments -->
	commentEnd      = "<!-- /gh-md:comment -->"
	newCommentStart = "<!-- gh-md:new-comment -->"
	newCommentEnd   = "<!-- /gh-md:new-comment -->"
)

// commentWithDepth holds a parsed comment and its indentation depth.
type commentWithDepth struct {
	comment  ParsedComment
	depth    int
	position int // position in original content for ordering
}

// parseComments extracts comments from the markdown content.
// It tracks indentation to determine parent-child relationships for replies.
// Supports both marker format (<!-- gh-md:comment -->) and simplified format (## New Comment, ## Reply).
func parseComments(content string) []ParsedComment {
	var commentsWithDepth []commentWithDepth

	// Parse existing marker-based comments
	commentsWithDepth = append(commentsWithDepth, parseMarkerComments(content)...)

	// Parse simplified format comments (## New Comment, ## Reply)
	commentsWithDepth = append(commentsWithDepth, parseNewCommentMarkers(content, commentsWithDepth)...)

	// Assign ParentID based on indentation hierarchy
	return assignParentIDs(commentsWithDepth)
}

// parseMarkerComments parses <!-- gh-md:comment --> blocks.
func parseMarkerComments(content string) []commentWithDepth {
	var result []commentWithDepth

	remaining := content
	offset := 0
	for {
		startIdx := strings.Index(remaining, commentStart)
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(remaining[startIdx:], commentEnd)
		if endIdx == -1 {
			break
		}
		endIdx += startIdx

		// Calculate indentation
		depth := calculateIndentation(content, offset+startIdx)

		// Parse this comment block
		block := remaining[startIdx : endIdx+len(commentEnd)]
		comment := parseCommentBlock(block)
		if comment != nil {
			result = append(result, commentWithDepth{
				comment:  *comment,
				depth:    depth,
				position: offset + startIdx,
			})
		}

		offset += endIdx + len(commentEnd)
		remaining = remaining[endIdx+len(commentEnd):]
	}

	return result
}

// parseNewCommentMarkers parses new comments within <!-- gh-md:new-comment --> markers.
// Supports reply_to attribute for discussion replies: <!-- gh-md:new-comment reply_to: DC_xxx -->
func parseNewCommentMarkers(content string, _ []commentWithDepth) []commentWithDepth {
	var result []commentWithDepth

	remaining := content
	for {
		// Find the new comment marker (with or without reply_to)
		startIdx := strings.Index(remaining, "<!-- gh-md:new-comment")
		if startIdx == -1 {
			break
		}

		// Find the end of the opening tag
		tagEndIdx := strings.Index(remaining[startIdx:], "-->")
		if tagEndIdx == -1 {
			break
		}
		tagEndIdx += startIdx

		// Find the closing tag
		closeIdx := strings.Index(remaining[tagEndIdx:], newCommentEnd)
		if closeIdx == -1 {
			break
		}
		closeIdx += tagEndIdx

		// Extract the opening tag to check for reply_to
		openingTag := remaining[startIdx : tagEndIdx+3]
		var parentID string
		if replyIdx := strings.Index(openingTag, "reply_to:"); replyIdx != -1 {
			// Extract the reply_to value
			afterReplyTo := openingTag[replyIdx+9:]
			endOfValue := strings.Index(afterReplyTo, "-->")
			if endOfValue == -1 {
				endOfValue = len(afterReplyTo)
			}
			parentID = strings.TrimSpace(afterReplyTo[:endOfValue])
		}

		// Extract body between tags
		bodyStart := tagEndIdx + 3 // after -->
		body := strings.TrimSpace(remaining[bodyStart:closeIdx])

		if body != "" {
			comment := ParsedComment{
				ID:       "", // New comment, no ID
				Body:     body,
				ParentID: parentID,
			}

			result = append(result, commentWithDepth{
				comment:  comment,
				depth:    0,
				position: startIdx * 1000,
			})
		}

		remaining = remaining[closeIdx+len(newCommentEnd):]
	}

	return result
}

// calculateIndentation returns the indentation depth based on leading whitespace.
func calculateIndentation(content string, pos int) int {
	// Find the start of the line containing pos
	lineStart := pos
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}

	// Count leading spaces (2 spaces = 1 depth level)
	spaces := 0
	for i := lineStart; i < pos; i++ {
		switch content[i] {
		case ' ':
			spaces++
		case '\t':
			spaces += 2
		}
	}

	return spaces / 2
}

// assignParentIDs assigns ParentID to comments based on depth hierarchy.
// If a comment already has ParentID set (from metadata), it's preserved.
func assignParentIDs(commentsWithDepth []commentWithDepth) []ParsedComment {
	var result []ParsedComment

	// Stack to track the most recent comment at each depth level
	// depthStack[depth] = ID of the most recent comment at that depth
	depthStack := make(map[int]string)

	for _, cwd := range commentsWithDepth {
		comment := cwd.comment
		depth := cwd.depth

		// If this comment is indented (depth > 0) and doesn't already have a parent, find its parent
		if depth > 0 && comment.ParentID == "" {
			// Look for the nearest comment at a lower depth
			for d := depth - 1; d >= 0; d-- {
				if parentID, ok := depthStack[d]; ok && parentID != "" {
					comment.ParentID = parentID
					break
				}
			}
		}

		// Update the stack: this comment becomes the most recent at its depth
		// Clear deeper levels since they can't be parents anymore
		if comment.ID != "" {
			depthStack[depth] = comment.ID
			for d := depth + 1; d < 10; d++ {
				delete(depthStack, d)
			}
		}

		result = append(result, comment)
	}

	return result
}

// parseCommentBlock parses a single comment block.
func parseCommentBlock(block string) *ParsedComment {
	// Extract metadata from the opening tag
	// Format: <!-- gh-md:comment\nid: xxx\nauthor: xxx\ncreated: xxx\n-->
	metaEndIdx := strings.Index(block, "-->")
	if metaEndIdx == -1 {
		return nil
	}

	metaSection := block[len(commentStart):metaEndIdx]
	metaLines := strings.Split(strings.TrimSpace(metaSection), "\n")

	var id, author, parentID string
	for _, line := range metaLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "id:") {
			id = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		} else if strings.HasPrefix(line, "author:") {
			author = strings.TrimSpace(strings.TrimPrefix(line, "author:"))
		} else if strings.HasPrefix(line, "parent:") {
			parentID = strings.TrimSpace(strings.TrimPrefix(line, "parent:"))
		}
	}

	// Extract body (content after metadata, before closing tag)
	bodyStart := metaEndIdx + 3 // Skip past -->
	bodyEndIdx := strings.Index(block, commentEnd)
	if bodyEndIdx == -1 {
		return nil
	}

	bodyContent := block[bodyStart:bodyEndIdx]

	// Remove the heading line (### @author (date))
	body := extractCommentBody(bodyContent)

	return &ParsedComment{
		ID:       id,
		Author:   author,
		Body:     body,
		ParentID: parentID,
	}
}

// extractCommentBody extracts the actual body from comment content,
// removing the heading line (### @author (date)).
func extractCommentBody(content string) string {
	content = strings.TrimSpace(content)
	lines := strings.Split(content, "\n")

	// Skip the heading line (starts with ### or ####)
	startIdx := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "###") || strings.HasPrefix(trimmed, "####") {
			startIdx = i + 1
			break
		}
	}

	if startIdx >= len(lines) {
		return ""
	}

	// Join remaining lines and trim
	body := strings.Join(lines[startIdx:], "\n")
	return strings.TrimSpace(body)
}

// WalkFilters specifies which files to include when walking.
type WalkFilters struct {
	Repo string // "owner/repo" format, empty = all repos
}

// WalkParsedFiles walks the gh-md root directory and calls the callback for each parsed file.
// Returns early if callback returns an error.
func WalkParsedFiles(filters WalkFilters, callback func(*ParsedFile) error) error {
	root, err := config.GetRootDir()
	if err != nil {
		return err
	}

	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}

		// Only process .md files
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Parse the file to extract metadata
		parsed, err := ParseFile(path)
		if err != nil {
			return nil // Skip files that can't be parsed
		}

		// Apply repo filter
		if filters.Repo != "" {
			repoPath := parsed.Owner + "/" + parsed.Repo
			if repoPath != filters.Repo {
				return nil
			}
		}

		return callback(parsed)
	})
}
