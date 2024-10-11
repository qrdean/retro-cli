package client

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type topicViewTest struct {
	focus  bool
	id     uint32
	header string
	list   list.Model
}

func (t topicViewTest) Init() tea.Cmd {
	return nil
}

func (t topicViewTest) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		t.list.SetSize(msg.Width-h, msg.Height-v)
	}
	t.list, cmd = t.list.Update(msg)
	return t, cmd
}

func (t topicViewTest) View() string {
	return docStyle.Render(t.list.View())
}

type model struct {
	TopicViewTest []topicViewTest
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var newitems []item
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "a" {
			newitem := item{title: "new item", desc: "new desc"}
			newitems = append(newitems, newitem)
		}
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd
	for i, topicView := range m.TopicViewTest {
		z, c := topicView.Update(msg)
		x, ok := z.(topicViewTest)
		if ok {
			for _, newitem := range newitems {
				x.list.InsertItem(-1, newitem)
			}
			m.TopicViewTest[i] = x
		}
		cmds = append(cmds, c)
	}
	cmd = tea.Batch(cmds...)
	// m.TopicViewTest[0], cmd = m.TopicViewTest[0].Update(msg)
	return m, cmd
}

func (m model) View() string {
	board := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.TopicViewTest[0].View(),
	)
	return docStyle.Render(board)
}

func RunTest() {
	items := []list.Item{
		item{title: "Raspberry Pi’s", desc: "I have ’em all over my house"},
		// item{title: "Nutella", desc: "It's good on toast"},
		// item{title: "Bitter melon", desc: "It cools you down"},
		// item{title: "Nice socks", desc: "And by that I mean socks without holes"},
		// item{title: "Eight hours of sleep", desc: "I had this once"},
		// item{title: "Cats", desc: "Usually"},
		// item{title: "Plantasia, the album", desc: "My plants love it too"},
		// item{title: "Pour over coffee", desc: "It takes forever to make though"},
		// item{title: "VR", desc: "Virtual reality...what is there to say?"},
		// item{title: "Noguchi Lamps", desc: "Such pleasing organic forms"},
		// item{title: "Linux", desc: "Pretty much the best OS"},
		// item{title: "Business school", desc: "Just kidding"},
		// item{title: "Pottery", desc: "Wet clay is a great feeling"},
		// item{title: "Shampoo", desc: "Nothing like clean hair"},
		// item{title: "Table tennis", desc: "It’s surprisingly exhausting"},
		// item{title: "Milk crates", desc: "Great for packing in your extra stuff"},
		// item{title: "Afternoon tea", desc: "Especially the tea sandwich part"},
		// item{title: "Stickers", desc: "The thicker the vinyl the better"},
		// item{title: "20° Weather", desc: "Celsius, not Fahrenheit"},
		// item{title: "Warm light", desc: "Like around 2700 Kelvin"},
		// item{title: "The vernal equinox", desc: "The autumnal equinox is pretty good too"},
		// item{title: "Gaffer’s tape", desc: "Basically sticky fabric"},
		// item{title: "Terrycloth", desc: "In other words, towel fabric"},
	}

	TopicViewTests := []topicViewTest{
		topicViewTest{focus: false, id: 0, header: "one", list: list.New(items, list.NewDefaultDelegate(), 0, 0)},
	}
	m := model{TopicViewTest: TopicViewTests}
	// m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	// m.list.Title = "My Fave Things"

	// p := tea.NewProgram(m, tea.WithAltScreen())
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
