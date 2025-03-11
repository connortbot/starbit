package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	TitleLeft TitleAlignment = iota
	TitleCenter
	TitleRight
)

func renderBoxBottom(width int) string {
	return "╰" + strings.Repeat("─", width-2) + "╯\n\n"
}

func renderMidline(width int) string {
	return "├" + strings.Repeat("─", width-2) + "┤\n"
}

// creates the top of a box with an optional title
func renderBoxTop(width int, title string, alignment TitleAlignment) string {
	if title == "" {
		return "╭" + strings.Repeat("─", width-2) + "╮"
	}

	titleLen := len(title) // account for space around the title
	leftWidth := 0
	rightWidth := 0

	switch alignment {
	case TitleLeft:
		leftWidth = 1
		rightWidth = width - titleLen - 5
	case TitleRight:
		leftWidth = width - titleLen - 5
		rightWidth = 1
	case TitleCenter:
		totalSpace := width - titleLen - 4
		leftWidth = totalSpace / 2
		rightWidth = totalSpace - leftWidth
	}

	return fmt.Sprintf("╭%s %s %s╮",
		strings.Repeat("─", leftWidth),
		title,
		strings.Repeat("─", rightWidth))
}

// padLine pads a string to fit in the box with borders
// If follow is true and content overflows, it will show the right side of the content
// If follow is false, it will truncate the content to fit
func padLine(content string, width int, follow bool) string {
	contentWidth := lipgloss.Width(content)
	padding := width - contentWidth - 3 // -3 for "│ " and "│"

	if padding < 0 {
		if follow {
			// show the right side of the content
			visibleWidth := width - 3 // -3 for "│ " and "│"
			visibleContent := content
			for lipgloss.Width(visibleContent) > visibleWidth {
				if len(visibleContent) > 1 {
					visibleContent = visibleContent[1:]
				} else {
					break
				}
			}
			return fmt.Sprintf("│ %s│", visibleContent)
		} else {
			// truncate the content
			truncated := content
			for lipgloss.Width(truncated) > width-3 {
				if len(truncated) > 1 {
					truncated = truncated[:len(truncated)-1]
				} else {
					break
				}
			}
			return fmt.Sprintf("│ %s│", truncated)
		}
	}
	return fmt.Sprintf("│ %s%s│", content, strings.Repeat(" ", padding))
}

// returns the minimum width a box would need to display the content
// OR
// the given width, whichever is greater
func calculateStretchBoxWidth(content string, title string, width int) int {
	// calculate content height by counting newlines
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")

	minWidth := 4                // minimum for borders and one space padding
	if len(title)+4 > minWidth { // title + spaces and borders
		minWidth = len(title) + 4
	}
	for _, line := range lines {
		lineWidth := lipgloss.Width(line) + 4 // +4 for borders and padding
		if lineWidth > minWidth {
			minWidth = lineWidth
		}
	}

	// use max of minimum width or provided width
	actualWidth := width
	if minWidth > width {
		actualWidth = minWidth
	}

	return actualWidth
}

// wrapInBox takes multi-line content and wraps it in a box with an optional title
func wrapInBox(content string, width int, title string, titleAlign TitleAlignment) string {
	// calculate content height by counting newlines
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	height := len(lines)
	
	// create viewport with full height to show all content
	viewport := NewScrollingViewport(
		content,
		calculateStretchBoxWidth(content, title, width),
		height,
		title,
		titleAlign,
	)
	return viewport.Render()
}

// sideBySide takes two pieces of content and places them side by side with optional padding between
func sideBySide(left, right string, padding int) string {
	leftLines := strings.Split(strings.TrimRight(left, "\n"), "\n")
	rightLines := strings.Split(strings.TrimRight(right, "\n"), "\n")

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	for len(leftLines) < maxLines {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, "")
	}

	var s strings.Builder
	paddingStr := strings.Repeat(" ", padding)

	for i := 0; i < maxLines; i++ {
		s.WriteString(leftLines[i])
		s.WriteString(paddingStr)
		s.WriteString(rightLines[i])
		s.WriteString("\n")
	}

	return s.String()
}

func sideBySideBoxes(padding int, boxes ...string) string {
	if len(boxes) == 0 {
		return ""
	}
	if len(boxes) == 1 {
		return boxes[0]
	}

	result := boxes[0]

	for i := 1; i < len(boxes); i++ {
		result = sideBySide(result, boxes[i], padding)
	}
	return result
}

func listBoxes(padding int, boxes ...string) string {
	if len(boxes) == 0 {
		return ""
	}
	var s strings.Builder
	for i, box := range boxes {
		s.WriteString(box)
		if i < len(boxes)-1 {
			s.WriteString(strings.Repeat("\n", padding))
		}
	}
	return s.String()
}
