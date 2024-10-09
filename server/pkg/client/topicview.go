package client

import (
	"fmt"
	"log"
	"pkg/shared"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type topicView struct {
	focus    bool
	id       uint32
	header   string
	stickies list.Model
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

func newTopicView(rawTopic shared.Topic) topicView {
	focus := false
	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	defaultList.SetShowHelp(false)
	return topicView{focus: focus, id: rawTopic.Id, header: string(rawTopic.Header[:]), stickies: defaultList}
}

func (t topicView) Init() tea.Cmd {
	return nil
}

func (t topicView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "v":
			i, ok := t.stickies.SelectedItem().(StickyItem)
			if ok {
				log.Printf("vote for %v\n", i.id)
			}
		}

		return t, cmd
	}

	t.stickies, cmd = t.stickies.Update(msg)
	return t, cmd
}

func (t topicView) View() string {
	return t.getStyle().Render(t.stickies.View())
}

func (t topicView) getStyle() lipgloss.Style {
	if t.Focused() {
		return lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Height(1).
			Width(1)
	}
	return lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		Height(1).
		Width(1)
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
				log.Println("here")
				return t.stickies.SetItem(idx, sticky)
			}
		}
	}
	log.Println("there")
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
	title					string
	description		string
}

func (s StickyItem) FilterValue() string {
	return s.stickyMessage
}

func (s StickyItem) Title() string {
	return s.stickyMessage
}

func (s StickyItem) Description() string {
	return fmt.Sprintf("Votes: %v", s.votes)
}

// func (s StickyItem) View() string {
// 	return fmt.Sprintf("something %v", s.id)
// }

func stickyItemFrom(stickyMsg shared.Sticky) StickyItem {
	return StickyItem{
		id:            stickyMsg.Id,
		posterId:      stickyMsg.PosterId,
		topicId:       stickyMsg.TopicId,
		votes:         stickyMsg.Votes,
		stickyMessage: string(stickyMsg.StickyMessage[:]),
		title: 				 string(stickyMsg.StickyMessage[:]),
		description: 	 fmt.Sprintf("Votes: %v", stickyMsg.Votes),
	}
}
