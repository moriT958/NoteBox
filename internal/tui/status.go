package tui

import (
	"time"

	"notebox/internal/tui/common"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// statusTTL is how long an error stays in the status line.
const statusTTL = 5 * time.Second

type clearStatusMsg struct{ seq int }

// Status is the bottom line: key help, or the latest error for a while.
type Status struct {
	com  *common.Common
	help help.Model
	err  string
	// seq identifies the latest error, so an older clear timer doesn't
	// clear a newer error.
	seq int
}

func NewStatus(com *common.Common) *Status {
	return &Status{com: com, help: help.New()}
}

func (s *Status) SetWidth(width int) {
	s.help.SetWidth(max(0, width-s.com.Styles.Help.GetHorizontalFrameSize()))
}

// SetError shows err and returns the command that clears it later.
func (s *Status) SetError(err error) tea.Cmd {
	s.seq++
	s.err = err.Error()
	seq := s.seq
	return tea.Tick(statusTTL, func(time.Time) tea.Msg { return clearStatusMsg{seq: seq} })
}

func (s *Status) Clear(seq int) {
	if seq == s.seq {
		s.err = ""
	}
}

func (s *Status) Render(width int, keys help.KeyMap) string {
	if s.err != "" {
		return s.com.Styles.Error.Width(width).Render(ansi.Truncate(s.err, width-1, "…"))
	}
	return s.com.Styles.Help.Width(width).Render(s.help.ShortHelpView(keys.ShortHelp()))
}

func (s *Status) RenderFullHelp(width int, keys help.KeyMap) string {
	return s.com.Styles.Help.Padding(1, 2).Width(width).Render(s.help.FullHelpView(keys.FullHelp()))
}
