package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	tabBarHeight = 2
	maxTabWidth  = 10
)

func (m *model) updatePreviewerSize(msg tea.WindowSizeMsg) {
	borderH, borderV := m.styles.BorderPassive.GetFrameSize()

	sidePanelWidth := msg.Width / layoutSidePanelRatio
	contentWidth := msg.Width - sidePanelWidth - borderH*2
	contentHeight := msg.Height - borderV - helpGuideHeight - tabBarHeight

	m.vp.SetWidth(contentWidth)
	m.vp.SetHeight(max(1, contentHeight))
}

func (m *model) updatePreviewerContent(msg renderPreviewMsg) tea.Cmd {
	m.vp.SetContent(string(msg))
	m.vp.GotoTop()
	return nil
}

func (m model) viewPreviewer() string {
	if m.vp.Width() == 0 {
		return ""
	}

	tabBar := m.buildTabBar()
	vpView := m.styles.Sized(m.vp.Width(), m.vp.Height()).Render(m.vp.View())

	var borderedContent string
	if m.focus == onPreviewer {
		borderedContent = m.styles.BorderActive.UnsetBorderTop().Render(vpView)
	} else {
		borderedContent = m.styles.BorderPassive.UnsetBorderTop().Render(vpView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, borderedContent)
}

func formatTabLabel(label string, width int) string {
	label = " " + label
	runes := []rune(label)
	if len(runes) >= width {
		return string(runes[:width-2]) + ".."
	}
	return label + strings.Repeat(" ", width-len(runes))
}

func (m model) buildTabBar() string {
	tabCount := len(m.tabs)
	if tabCount == 0 || m.vp.Width() == 0 {
		return ""
	}

	totalWidth := m.vp.Width() + 2
	tabOuterWidth := min(totalWidth/tabCount, maxTabWidth)
	tabInnerWidth := tabOuterWidth - 2
	tabsTotal := tabOuterWidth * tabCount
	remaining := totalWidth - tabsTotal

	var topsRow, contentRow strings.Builder

	for i, label := range m.tabs {
		colorStyle := m.styles.PassiveColor
		if i == m.activeTab {
			colorStyle = m.styles.ActiveColor
		}
		topsRow.WriteString(colorStyle.Render("┏" + strings.Repeat("━", tabInnerWidth) + "┓"))
		contentRow.WriteString(colorStyle.Render("┃" + formatTabLabel(label, tabInnerWidth) + "┃"))
	}

	dividerRow := m.buildDividerRow(tabCount, tabInnerWidth, remaining)

	return topsRow.String() + "\n" + contentRow.String() + "\n" + m.styles.PassiveColor.Render(dividerRow)
}

func (m model) buildDividerRow(tabCount, tabInnerWidth, remaining int) string {
	var divider strings.Builder

	if m.activeTab == 0 {
		divider.WriteRune('┃')
	} else {
		divider.WriteRune('┣')
	}

	for i := range m.tabs {
		isActive := i == m.activeTab
		isLast := i == tabCount-1

		if isActive {
			divider.WriteString(strings.Repeat(" ", tabInnerWidth))
		} else {
			divider.WriteString(strings.Repeat("━", tabInnerWidth))
		}

		if isLast {
			if remaining > 0 {
				if isActive {
					divider.WriteRune('┗')
				} else {
					divider.WriteRune('┻')
				}
				divider.WriteString(strings.Repeat("━", remaining-1))
				divider.WriteRune('┓')
			} else {
				if isActive {
					divider.WriteRune('┃')
				} else {
					divider.WriteRune('┫')
				}
			}
		} else {
			nextActive := (i + 1) == m.activeTab
			switch {
			case !isActive && nextActive:
				divider.WriteString("┻┛")
			case isActive && !nextActive:
				divider.WriteString("┗┻")
			default:
				divider.WriteString("┻┻")
			}
		}
	}

	return divider.String()
}
