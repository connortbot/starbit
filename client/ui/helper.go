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

func renderBoxBottom(width int, style *lipgloss.Style) string {
	bottom := "╰" + strings.Repeat("─", width-2) + "╯\n\n"
	if style != nil {
		return style.Render(bottom)
	}
	return bottom
}

func renderMidline(width int, style *lipgloss.Style) string {
	midline := "├" + strings.Repeat("─", width-2) + "┤\n"
	if style != nil {
		return style.Render(midline)
	}
	return midline
}

// creates the top of a box with an optional title
func renderBoxTop(width int, title string, alignment TitleAlignment, style *lipgloss.Style) string {
	var top string
	if title == "" {
		top = "╭" + strings.Repeat("─", width-2) + "╮"
	} else {
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

		top = fmt.Sprintf("╭%s %s %s╮",
			strings.Repeat("─", leftWidth),
			title,
			strings.Repeat("─", rightWidth))
	}

	if style != nil {
		return style.Render(top)
	}
	return top
}

// padLine pads a string to fit in the box with borders
// If follow is true and content overflows, it will show the right side of the content
// If follow is false, it will truncate the content to fit
func padLine(content string, width int, follow bool, style *lipgloss.Style) string {
	contentWidth := lipgloss.Width(content)
	padding := width - contentWidth - 3 // -3 for "│ " and "│"

	var line string

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
			line = fmt.Sprintf("│ %s│", visibleContent)
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
			line = fmt.Sprintf("│ %s│", truncated)
		}
	} else {
		line = fmt.Sprintf("│ %s%s│", content, strings.Repeat(" ", padding))
	}

	if style != nil {
		// we only want to style the border chars, not the content
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			// style just the first border char
			styled := style.Render(parts[0])
			// get the content and padding
			contentParts := strings.SplitN(parts[1], "│", 2)
			if len(contentParts) == 2 {
				// style just the last border char
				styledEnd := style.Render("│")
				line = styled + " " + contentParts[0] + styledEnd
			} else {
				line = styled + " " + parts[1]
			}
		}
	}

	return line
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
func wrapInBox(content string, width int, height int, title string, titleAlign TitleAlignment, style *lipgloss.Style) string {
	if height <= 0 {
		height = len(strings.Split(strings.TrimRight(content, "\n"), "\n"))
	}

	viewport := NewScrollingViewport(
		content,
		calculateStretchBoxWidth(content, title, width),
		height,
		title,
		titleAlign,
	)

	if style != nil {
		viewport.SetStyle(style)
	}

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

	leftWidth := 0
	for _, line := range leftLines {
		lineWidth := lipgloss.Width(line)
		if lineWidth > leftWidth {
			leftWidth = lineWidth
		}
	}

	for len(leftLines) < maxLines {
		leftLines = append(leftLines, strings.Repeat(" ", leftWidth))
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, "")
	}

	var s strings.Builder
	paddingStr := strings.Repeat(" ", padding)

	for i := 0; i < maxLines; i++ {
		// ensure the left stuff is padded to a consistent width
		lineWidth := lipgloss.Width(leftLines[i])
		extraPadding := leftWidth - lineWidth

		s.WriteString(leftLines[i])
		if extraPadding > 0 {
			s.WriteString(strings.Repeat(" ", extraPadding))
		}
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
