package client

import (
	"fmt"
	"pkg/shared"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type topicView struct {
	focus              bool
	id                 uint32
	header             string
	currentItemId      uint32
	stickies           list.Model
	height             int
	width              int
	parentWidthContext int
}

func (t *topicView) Focus() {
	t.focus = true
}

func (t *topicView) Blur() {
	t.focus = false
}

func (t *topicView) Focused() bool {
	return t.focus
}

func newTestTopicViewWithList(id uint32, header [255]byte, stickyItem StickyItem) topicView {
	focus := false
	delegate := list.NewDefaultDelegate()
	defaultList := list.New([]list.Item{stickyItem}, delegate, 0, 0)
	defaultList.SetShowHelp(false)
	defaultList.SetWidth(25)
	defaultList.SetHeight(35)
	return topicView{focus: focus, id: id, header: string(header[:]), stickies: defaultList}
}

func newTopicView(rawTopic shared.Topic) topicView {
	focus := false
	delegate := list.NewDefaultDelegate()
	defaultList := list.New([]list.Item{}, delegate, 0, 0)
	defaultList.SetShowHelp(false)
	defaultList.SetWidth(25)
	defaultList.SetHeight(35)
	return topicView{focus: focus, id: rawTopic.Id, header: string(rawTopic.Header[:]), stickies: defaultList}
}

func (t topicView) Init() tea.Cmd {
	return nil
}

func (t topicView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		_, v := t.getStyle().GetFrameSize()
		t.stickies.SetWidth(t.parentWidthContext + 5)
		t.stickies.SetHeight(msg.Height - v - 2)
	}

	t.stickies.SetWidth(t.parentWidthContext + 5)
	t.stickies, cmd = t.stickies.Update(msg)
	if t.stickies.SelectedItem() != nil {
		item := t.stickies.SelectedItem().(StickyItem)
		t.currentItemId = item.id
	}
	return t, cmd
}

func (t topicView) View() string {
	return t.getStyle().
		Render(t.header + "\n\n" + t.stickies.View())
}

func (t topicView) getStyle() lipgloss.Style {
	if t.Focused() {
		return lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			// Height(t.height)
			Height(1)
		// Width(t.parentWidthContext - 2)
		// Width(0)

	}
	return lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		// Height(t.height)
		Height(1)
	// Width(t.parentWidthContext - 2)
	// Width(0)
}

func (t *topicView) updateAllSticky(sticky []StickyItem) {
	for _, stickyItem := range sticky {
		for idx, item := range t.stickies.Items() {
			i, ok := item.(StickyItem)
			if ok {
				if i.id == stickyItem.id {
					t.stickies.SetItem(idx, stickyItem)
					break
				}
			}
		}
	}
}

func (t *topicView) findAndUpdateSticky(sticky StickyItem) tea.Cmd {
	for idx, item := range t.stickies.Items() {
		i, ok := item.(StickyItem)
		if ok {
			if i.id == sticky.id {
				return t.stickies.SetItem(idx, sticky)
			}
		}
	}
	return t.appendSticky(sticky)
}

func (t *topicView) updateSticky(i int, sticky StickyItem) tea.Cmd {
	return t.stickies.SetItem(i, sticky)
}

func (t *topicView) appendSticky(sticky StickyItem) tea.Cmd {
	return t.stickies.InsertItem(-1, sticky)
}

type StickyItem struct {
	id            uint32
	posterId      uint32
	topicId       uint32
	votes         uint32
	stickyMessage string
	title         string
	description   string
	width         int
}

func (s StickyItem) FilterValue() string {
	return s.title
}

func (s StickyItem) Title() string {
	display := strings.TrimSpace(s.title)
	// if len(display) > 130 {
	// 	display = fmt.Sprintf("%s ...", s.title[:120])
	// }
	return lipgloss.NewStyle().Render(display)
	// return s.title
}

func (s StickyItem) Description() string {
	return s.description
}

// func (s StickyItem) View() string {
// 	return fmt.Sprintf("something %v", s.id)
// }

func testStickyItem(id, posterId, topicId, votes uint32, stickyMessage [255]byte) StickyItem {
	return StickyItem{
		id:            id,
		posterId:      posterId,
		topicId:       topicId,
		votes:         votes,
		stickyMessage: string(stickyMessage[:]),
		title:         string(stickyMessage[:]),
		description:   fmt.Sprintf("QD - Votes: %v", votes),
	}
}

func stickyItemFrom(stickyMsg shared.Sticky) StickyItem {
	return StickyItem{
		id:            stickyMsg.Id,
		posterId:      stickyMsg.PosterId,
		topicId:       stickyMsg.TopicId,
		votes:         stickyMsg.Votes,
		stickyMessage: string(stickyMsg.StickyMessage[:]),
		title:         string(stickyMsg.StickyMessage[:]),
		description:   fmt.Sprintf("QD - Votes: %v", stickyMsg.Votes),
	}
}
