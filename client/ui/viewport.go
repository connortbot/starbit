package ui

import (
	"strings"
)

type ScrollingViewport struct {
	content    string
	width      int
	height     int
	scrollY    int
	title      string
	titleAlign TitleAlignment
}

func NewScrollingViewport(content string, width, height int, title string, titleAlign TitleAlignment) *ScrollingViewport {
	return &ScrollingViewport{
		content:    content,
		width:      width,
		height:     height,
		scrollY:    0,
		title:      title,
		titleAlign: titleAlign,
	}
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
	s.WriteString(renderBoxTop(v.width, v.title, v.titleAlign) + "\n")
	lines := strings.Split(strings.TrimRight(v.content, "\n"), "\n")

	startLine := v.scrollY
	endLine := startLine + v.height
	if endLine > len(lines) {
		endLine = len(lines)
	}

	for i := startLine; i < endLine; i++ {
		s.WriteString(padLine(lines[i], v.width, false) + "\n")
	}

	for i := endLine - startLine; i < v.height; i++ {
		s.WriteString(padLine("", v.width, false) + "\n")
	}

	s.WriteString(renderBoxBottom(v.width))
	return s.String()
}
