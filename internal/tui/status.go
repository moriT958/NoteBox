package tui

import (
	"time"

	"notebox/internal/tui/internal/common"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// How long a message stays in the status line.
const (
	errorTTL = 5 * time.Second
	infoTTL  = 2 * time.Second
)

type clearStatusMsg struct{ seq int }

// status is the bottom line: key help, or the latest message (an error or
// a notice) for a while.
type status struct {
	com   *common.Common
	help  help.Model
	text  string
	style lipgloss.Style
	// seq identifies the latest message, so an older clear timer doesn't
	// clear a newer message.
	seq int
}

func newStatus(com *common.Common) *status {
	return &status{com: com, help: help.New()}
}

func (s *status) SetWidth(width int) {
	s.help.SetWidth(max(0, width-s.com.Styles.Help.GetHorizontalFrameSize()))
}

// SetError shows err and returns the command that clears it later.
func (s *status) SetError(err error) tea.Cmd {
	return s.set(err.Error(), s.com.Styles.Error, errorTTL)
}

// SetInfo shows a notice and returns the command that clears it later.
func (s *status) SetInfo(text string) tea.Cmd {
	return s.set(text, s.com.Styles.Info, infoTTL)
}

func (s *status) set(text string, style lipgloss.Style, ttl time.Duration) tea.Cmd {
	s.seq++
	s.text, s.style = text, style
	seq := s.seq
	return tea.Tick(ttl, func(time.Time) tea.Msg { return clearStatusMsg{seq: seq} })
}

func (s *status) Clear(seq int) {
	if seq == s.seq {
		s.text = ""
	}
}

func (s *status) Render(width int, keys help.KeyMap) string {
	if s.text != "" {
		return s.style.Width(width).Render(ansi.Truncate(s.text, width-1, "…"))
	}
	return s.com.Styles.Help.Width(width).Render(s.help.ShortHelpView(keys.ShortHelp()))
}

func (s *status) RenderFullHelp(width int, keys help.KeyMap) string {
	return s.com.Styles.Help.Padding(1, 2).Width(width).Render(s.help.FullHelpView(keys.FullHelp()))
}
