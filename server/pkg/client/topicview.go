package client

import (
	"pkg/shared"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

func (t *topicView) updateSticky(i int, sticky StickyItem) tea.Cmd {
	return t.stickies.SetItem(i, sticky)
}

func (t *topicView) appendSticky(i int, sticky StickyItem) tea.Cmd{
	return t.stickies.InsertItem(-1, sticky)
}

type StickyItem struct {
	id            uint32
	posterId      uint32
	topicId       uint32
	votes         uint32
	stickyMessage string
}

func (s StickyItem) FilterValue() string {
	return s.stickyMessage
}
