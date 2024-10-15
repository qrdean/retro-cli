package client

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"pkg/shared"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var appStyle = lipgloss.NewStyle().Margin(1, 2, 0, 2)

type ViewTopic struct {
	Topic    shared.Topic
	Stickies []shared.Sticky
}

type size int

const (
	undersized size = iota
	small
	medium
	large
)

type Model struct {
	Connection      net.Conn
	SomeThing       string
	Topics          []shared.Topic
	Stickies        []shared.Sticky
	ViewTopic       map[int]ViewTopic
	TopicViews      map[uint32]topicView
	PointToSticky   uint32
	ErrorMsg        string
	CurrentTopic    uint32
	textinput       textinput.Model
	textMode        bool
	selectedMode    bool
	widthContainer  int
	heightContainer int
	viewportHeight  int
	viewportWidth   int
	widthContent    int
	TopicAck        bool
	TopicNumber     uint32
	ViewReady       bool
	size            size
}

func initialModel(conn net.Conn) Model {
	ti := textinput.New()
	ti.Placeholder = "New Sticky"
	ti.Focus()
	ti.CharLimit = 250
	ti.Width = 128

	return Model{
		Connection:    conn,
		SomeThing:     "hello",
		Topics:        []shared.Topic{},
		Stickies:      []shared.Sticky{},
		TopicViews:    make(map[uint32]topicView),
		PointToSticky: 0,
		ErrorMsg:      "",
		CurrentTopic:  0,
		textinput:     ti,
		TopicAck:      false,
	}
}

func (m Model) Init() tea.Cmd {
	// testInitialTopics(m)
	return initMessageHandler(m.Connection)
}

func testInitialTopics(m Model) {
	var bytes [255]byte
	stringThing := []byte("Topic 0 World")
	if len(stringThing) <= 255 {
		copy(bytes[:len(stringThing)], stringThing)
	} else {
		log.Printf("too long %v", len(stringThing))
		copy(bytes[:], stringThing)
	}
	m.TopicViews[0] = newTestTopicViewWithList(0, bytes, testStickyItem(1, 1, 1, 2, bytes))
	tp := m.TopicViews[0]
	tp.Focus()
	m.TopicViews[0] = tp

	stringThing = []byte("Topic 1 World")
	if len(stringThing) <= 255 {
		copy(bytes[:len(stringThing)], stringThing)
	} else {
		log.Printf("too long %v", len(stringThing))
		copy(bytes[:], stringThing)
	}
	m.TopicViews[1] = newTestTopicViewWithList(1, bytes, testStickyItem(1, 1, 1, 2, bytes))

	stringThing = []byte("Topic 2 World")
	if len(stringThing) <= 255 {
		copy(bytes[:len(stringThing)], stringThing)
	} else {
		log.Printf("too long %v", len(stringThing))
		copy(bytes[:], stringThing)
	}

	m.TopicViews[2] = newTestTopicViewWithList(2, bytes, testStickyItem(1, 1, 1, 2, bytes))
	topicView := m.TopicViews[0]

	stringThing = []byte("Hello World")
	if len(stringThing) <= 255 {
		copy(bytes[:len(stringThing)], stringThing)
	} else {
		log.Printf("too long %v", len(stringThing))
		copy(bytes[:], stringThing)
	}
	// topicView.findAndUpdateSticky(testStickyItem(1, 1, 1, 2, bytes))
	topicView.appendSticky(testStickyItem(1, 1, 1, 2, bytes))
}

type ParentWidthChange struct {
	width int
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	var stickies []StickyItem
	dontupdate := false

	if m.textMode {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				somevalue := m.textinput.Value()

				msg := somevalue
				sticky, err := shared.NewAddSticky(1, m.CurrentTopic, msg)
				if err != nil {
					// log.Println(err)
					// continue
				}
				var stickyBytes shared.AddStickyBytes
				stickyBytes = sticky.MarshalBinary()
				_, err = stickyBytes.WriteTo(m.Connection)
				if err != nil {
					m.ErrorMsg = string(err.Error())
				}

				m.textMode = false
				m.textinput.Reset()
			}
		}

		m.textinput, cmd = m.textinput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewportHeight = msg.Height
		m.viewportWidth = msg.Width
		switch {
		case m.viewportHeight < 10 || m.viewportWidth < 20:
			m.size = undersized
			m.widthContainer = m.viewportWidth
			m.heightContainer = m.viewportHeight

		case m.viewportWidth < 40:
			m.size = small
			m.widthContainer = m.viewportWidth
			m.heightContainer = m.viewportHeight

		case m.viewportWidth < 60:
			m.size = medium
			m.widthContainer = 40
			m.heightContainer = int(math.Min(float64(msg.Height), 30))

		default:
			m.size = large
			m.widthContainer = 100
			m.heightContainer = int(math.Min(float64(msg.Height), 30))
		}

		m.widthContent = m.widthContainer - 4

	case ErrMsg:
		m.ErrorMsg = string(msg.err.Error())

	case Break:
		quit := shared.NewQuit(1)
		var quitBytes shared.QuitBytes = quit.MarshalBinary()
		_, err := quitBytes.WriteTo(m.Connection)
		if err != nil {
			m.ErrorMsg = string(err.Error())
		}
		return m, tea.Quit

	case TopicLength:
		m.TopicNumber = uint32(msg)
		_, err := m.Connection.Write([]byte{1, 42})
		if err != nil {
			fmt.Println(err)
		}

	case shared.Sticky:
		stickies = append(stickies, stickyItemFrom(msg))

	case shared.Topic:
		m.TopicViews[msg.Id] = newTopicView(msg)
		fmt.Println("length of topics is", len(m.TopicViews))
		if !m.TopicAck {
			m.TopicAck = true
			tp := m.TopicViews[msg.Id]
			tp.Focus()
			m.TopicViews[msg.Id] = tp
		}

		if len(m.TopicViews) == int(m.TopicNumber) {
			_, err := m.Connection.Write([]byte{1, 41})
			if err != nil {
				fmt.Println(err)
				m.ErrorMsg = string(err.Error())
			}
			m.ViewReady = true
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			quit := shared.NewQuit(1)
			var quitBytes shared.QuitBytes = quit.MarshalBinary()
			_, err := quitBytes.WriteTo(m.Connection)
			if err != nil {
				// fmt.Println(err)
				m.ErrorMsg = string(err.Error())
			}
			// fmt.Println("wrote %v bytes %v\n", n, quitBytes[:])
			return m, tea.Quit

		case "n":
			m.textinput, cmd = m.textinput.Update(nil)
			m.textMode = true
			return m, cmd

		case "s":
			m.selectedMode = true

		case "b":
			m.selectedMode = false

		case "v":
			tp := m.TopicViews[m.CurrentTopic]
			voteSticky := shared.NewVoteSticky(tp.currentItemId)
			var voteBytes shared.VoteBytes = voteSticky.MarshalBinary()
			_, err := voteBytes.WriteTo(m.Connection)
			if err != nil {
				// fmt.Println(err)
				m.ErrorMsg = string(err.Error())
			}

		case "tab":
			tp := m.TopicViews[m.CurrentTopic]
			tp.Blur()
			m.TopicViews[m.CurrentTopic] = tp
			m.CurrentTopic++
			if m.CurrentTopic >= uint32(len(m.TopicViews)) {
				m.CurrentTopic = 0
			}
			tp = m.TopicViews[m.CurrentTopic]
			tp.Focus()
			m.TopicViews[m.CurrentTopic] = tp

		case "j", "k":
			tp := m.TopicViews[m.CurrentTopic]
			z, c := tp.Update(msg)
			x, ok := z.(topicView)
			if ok {
				m.TopicViews[m.CurrentTopic] = x
				cmds = append(cmds, c)
			}
			dontupdate = true
		}
	}

	if !dontupdate {
		for i, tv := range m.TopicViews {
			if m.selectedMode {
				tv.parentWidthContext = m.widthContent
			} else {
				tv.parentWidthContext = m.widthContent / len(m.TopicViews)
			}
			z, c := tv.Update(msg)
			x, ok := z.(topicView)
			if ok {
				for _, newitem := range stickies {
					if i == newitem.topicId {
						s := x.findAndUpdateSticky(newitem)
						cmds = append(cmds, s)
					}
				}
				m.TopicViews[i] = x
			}
			cmds = append(cmds, c)
		}
	}
	cmd = tea.Batch(cmds...)
	return m, cmd
}

func (m Model) View() string {

	if !m.ViewReady {
		return "Loading..."
	}

	var board string
	if m.selectedMode {
		board = m.TopicViews[m.CurrentTopic].View()
	} else {
		// board = lipgloss.JoinHorizontal(
		// 	lipgloss.Left,
		// 	m.TopicViews[0].View(),
		// 	m.TopicViews[1].View(),
		// 	m.TopicViews[2].View(),
		// 	m.TopicViews[3].View(),
		// 	m.TopicViews[4].View(),
		// )
		board = getBoardUpdated(m)
	}
	if m.textMode {
		textView := "Adding to: "
		textView += m.TopicViews[m.CurrentTopic].header
		textView += "\n"
		textView += m.textinput.View()
		return appStyle.Render(textView)
	}

	// return appStyle.Render(board)
	return appStyle.Render(board)
}

func getBoardUpdated(m Model) string {
	var view string
	var topicNumbers []string
	for i := range m.TopicNumber {
		topicNumbers = append(topicNumbers, m.TopicViews[i].View())
	}
	view = lipgloss.JoinHorizontal(
		lipgloss.Left,
		topicNumbers...,
	)
	return view
}

func getBoard(m Model) string {
	var view string
	switch m.TopicNumber {
	case 1:
		view = lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.TopicViews[0].View(),
		)
	case 2:
		view = lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.TopicViews[0].View(),
			m.TopicViews[1].View(),
		)
	case 3:
		view = lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.TopicViews[0].View(),
			m.TopicViews[1].View(),
			m.TopicViews[2].View(),
		)
	case 4:
		view = lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.TopicViews[0].View(),
			m.TopicViews[1].View(),
			m.TopicViews[2].View(),
			m.TopicViews[3].View(),
		)
	case 5:
		view = lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.TopicViews[0].View(),
			m.TopicViews[1].View(),
			m.TopicViews[2].View(),
			m.TopicViews[3].View(),
			m.TopicViews[4].View(),
		)
	}

	return view

}

func RunTui() {
	addr := "127.0.0.1:49000"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	p := tea.NewProgram(initialModel(conn))

	go func() {
		// newReader := bufio.NewReader(conn)
		newReader := bufio.NewReaderSize(conn, 4096*100)
		for {
			data := refactorHandleMessage(newReader)
			p.Send(data)
		}
	}()

	_, err = conn.Write([]byte{1, 40})
	if err != nil {
		fmt.Println(err)
	}
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, theres been an error: %v", err)
		os.Exit(1)
	}
}
