package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type ScrollingViewport struct {
	content    string
	width      int
	height     int
	scrollY    int
	title      string
	titleAlign TitleAlignment
	style      *lipgloss.Style
}

func NewScrollingViewport(content string, width, height int, title string, titleAlign TitleAlignment) *ScrollingViewport {
	return &ScrollingViewport{
		content:    content,
		width:      width,
		height:     height,
		scrollY:    0,
		title:      title,
		titleAlign: titleAlign,
		style:      nil,
	}
}

func (v *ScrollingViewport) SetStyle(style *lipgloss.Style) {
	v.style = style
}

func (v *ScrollingViewport) ScrollDown() {
	lines := strings.Split(strings.TrimRight(v.content, "\n"), "\n")
	if v.scrollY < len(lines)-v.height {
		v.scrollY++
	}
}

func (v *ScrollingViewport) ScrollUp() {
	if v.scrollY > 0 {
		v.scrollY--
	}
}

func (v *ScrollingViewport) ScrollToTop() {
	v.scrollY = 0
}

func (v *ScrollingViewport) ScrollToBottom() {
	lines := strings.Split(strings.TrimRight(v.content, "\n"), "\n")
	v.scrollY = len(lines) - v.height
	if v.scrollY < 0 {
		v.scrollY = 0
	}
}

func (v *ScrollingViewport) UpdateContent(content string) {
	v.content = content
}

func (v *ScrollingViewport) Render() string {
	var s strings.Builder
	s.WriteString(renderBoxTop(v.width, v.title, v.titleAlign, v.style) + "\n")
	lines := strings.Split(strings.TrimRight(v.content, "\n"), "\n")

	startLine := v.scrollY
	endLine := startLine + v.height
	if endLine > len(lines) {
		endLine = len(lines)
	}

	for i := startLine; i < endLine; i++ {
		s.WriteString(padLine(lines[i], v.width, false, v.style) + "\n")
	}

	for i := endLine - startLine; i < v.height; i++ {
		s.WriteString(padLine("", v.width, false, v.style) + "\n")
	}

	s.WriteString(renderBoxBottom(v.width, v.style))
	return s.String()
}
