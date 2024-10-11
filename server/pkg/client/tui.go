package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"pkg/shared"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var appStyle = lipgloss.NewStyle().Margin(1, 2, 0, 2)

type ViewTopic struct {
	Topic    shared.Topic
	Stickies []shared.Sticky
}

type Model struct {
	Connection    net.Conn
	SomeThing     string
	Topics        []shared.Topic
	Stickies      []shared.Sticky
	ViewTopic     map[int]ViewTopic
	TopicViews    map[uint32]topicView
	PointToSticky uint32
	ErrorMsg      string
	CurrentTopic  uint32
}

func initialModel(conn net.Conn) Model {
	return Model{
		Connection:    conn,
		SomeThing:     "hello",
		Topics:        []shared.Topic{},
		Stickies:      []shared.Sticky{},
		TopicViews:    make(map[uint32]topicView),
		PointToSticky: 0,
		ErrorMsg:      "",
		CurrentTopic:  0,
	}
}

func (m Model) Init() tea.Cmd {
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
	return initMessageHandler(m.Connection)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	var stickies []StickyItem
	dontupdate := false
	switch msg := msg.(type) {
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

	case shared.Sticky:
		stickies = append(stickies, stickyItemFrom(msg))

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

		case "a":
			msg := "Hello Msg"
			sticky, err := shared.NewAddSticky(1, 1, msg)
			if err != nil {
				// log.Println(err)
				// continue
			}
			var stickyBytes shared.AddStickyBytes
			stickyBytes = sticky.MarshalBinary()
			_, err = stickyBytes.WriteTo(m.Connection)
			if err != nil {
				// fmt.Println(err)
				m.ErrorMsg = string(err.Error())
			}
			// fmt.Printf("successfully wrote %v bytes %v\n", n, stickyBytes[:n-8])
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

func (m Model) Update2(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	// case Init:
	// 	fmt.Println("init")
	// 	fmt.Println(msg)
	// 	return m, refactorHandleMessage(m.Reader)
	case ErrMsg:
		// fmt.Println(msg.err.Error())
		m.ErrorMsg = string(msg.err.Error())
		return m, nil // refactorHandleMessage(m.Connection)
	case Break:
		// fmt.Println("breaking")
		quit := shared.NewQuit(1)
		var quitBytes shared.QuitBytes = quit.MarshalBinary()
		_, err := quitBytes.WriteTo(m.Connection)
		if err != nil {
			// fmt.Println(err)
			m.ErrorMsg = string(err.Error())
		}
		// fmt.Println("wrote %v bytes %v\n", n, quitBytes[:])
		return m, tea.Quit

	case shared.Sticky:
		// fmt.Println("sticky")
		// fmt.Println(msg.Id)
		for idx, sticky := range m.Stickies {
			if sticky.Id == msg.Id {
				m.Stickies[idx] = msg
				// log.Println("handling sticky exists")
				return m, nil //refactorHandleMessage(m.Connection)
			}
		}

		topic, ok := m.TopicViews[msg.TopicId]
		var cmd tea.Cmd
		if ok {
			cmd = topic.findAndUpdateSticky(stickyItemFrom(msg))
		}
		// log.Println("handling sticky add")
		m.Stickies = append(m.Stickies, msg)
		// m.TopicViews[0].Update(msg)
		// m.TopicViews[1].Update(msg)
		// m.TopicViews[2].Update(msg)
		return m, cmd //refactorHandleMessage(m.Connection)

	case shared.Topic:
		// fmt.Println("sticky")
		// fmt.Println(msg.Id)
		for idx, topic := range m.Topics {
			if topic.Id == msg.Id {
				m.Topics[idx] = msg
				return m, nil // refactorHandleMessage(m.Connection)
			}
		}

		m.TopicViews[msg.Id] = newTopicView(msg)
		m.Topics = append(m.Topics, msg)
		// m.TopicViews[0].Update(msg)
		// m.TopicViews[1].Update(msg)
		// m.TopicViews[2].Update(msg)
		// return m, nil //refactorHandleMessage(m.Connection)

	case shared.Pointer:
		// fmt.Println("sticky")
		// fmt.Println(msg.PointerId)
		m.PointToSticky = msg.PointerId
		// m.TopicViews[0].Update(msg)
		// m.TopicViews[1].Update(msg)
		// m.TopicViews[2].Update(msg)
		// return m, nil // refactorHandleMessage(m.Connection)

	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			msg := "Hello Msg"
			sticky, err := shared.NewAddSticky(1, 1, msg)
			if err != nil {
				// log.Println(err)
				// continue
			}
			var stickyBytes shared.AddStickyBytes
			stickyBytes = sticky.MarshalBinary()
			_, err = stickyBytes.WriteTo(m.Connection)
			if err != nil {
				// fmt.Println(err)
				m.ErrorMsg = string(err.Error())
			}
			// fmt.Printf("successfully wrote %v bytes %v\n", n, stickyBytes[:n-8])
			return m, nil // refactorHandleMessage(m.Connection)
		case "v":
			// println(msg)
			// int, err := strconv.Atoi(msg)
			// if err != nil {
			// 	log.Println(err)
			// 	continue
			// }
			voteSticky := shared.NewVoteSticky(uint32(1))
			var voteBytes shared.VoteBytes = voteSticky.MarshalBinary()
			_, err := voteBytes.WriteTo(m.Connection)
			if err != nil {
				// fmt.Println(err)
				m.ErrorMsg = string(err.Error())
			}

			// return m, nil
			// fmt.Printf("wrote %v bytes %v\n", n, voteBytes[:])
		case "p":
			pointTo := shared.NewPointToSticky(uint32(1))
			var pointToBytes shared.PointToStickyBytes = pointTo.MarshalBinary()
			_, err := pointToBytes.WriteTo(m.Connection)
			if err != nil {
				// fmt.Println(err)
				m.ErrorMsg = string(err.Error())
			}
			// return m, nil
			// fmt.Printf("wrote %v bytes %v\n", n, pointToBytes[:])

		// case "tab":
		// 	m.TopicViews[curr]

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
		}
	}

	var cmds []tea.Cmd
	for i, tv := range m.TopicViews {
		z, c := tv.Update(msg)
		x, ok := z.(topicView)
		if ok {
			m.TopicViews[i] = x
		}
		cmds = append(cmds, c)
	}
	cmd = tea.Batch(cmds...)

	return m, cmd //refactorHandleMessage(m.Connection)
}

func (m Model) View() string {
	// var s string
	// s += fmt.Sprintf("length of stickies: %v and length of topics %v\n", len(m.Stickies), len(m.Topics))
	// mapstring := make(map[uint32]ViewTopic)
	// s += "\n"
	// if m.ErrorMsg != "" {
	// 	s += fmt.Sprintf("error msg: %v\n", m.ErrorMsg)
	// }
	//
	// for _, topic := range m.Topics {
	// 	newTopic := ViewTopic{Topic: topic}
	// 	// news := ""
	// 	// news += string(topic.Header[:])
	// 	// news += "\n"
	// 	mapstring[topic.Id] = newTopic
	// }
	//
	// for _, sticky := range m.Stickies {
	// 	viewTopic := mapstring[sticky.TopicId]
	// 	viewTopic.Stickies = append(viewTopic.Stickies, sticky)
	// 	mapstring[sticky.TopicId] = viewTopic
	// }
	//
	// for _, viewTopic := range mapstring {
	// 	s += fmt.Sprintf("Topic Name: %v\n", string(viewTopic.Topic.Header[:]))
	// 	for _, sticky := range viewTopic.Stickies {
	// 		s += fmt.Sprintf("%v votes: %v\n", string(sticky.StickyMessage[:]), sticky.Votes)
	// 	}
	// 	s += "\n"
	// }

	board := lipgloss.JoinHorizontal(
		lipgloss.Left,
		// topicViewString,
		m.TopicViews[0].View(),
		m.TopicViews[1].View(),
		m.TopicViews[2].View(),
	)
	// s += "press q to quit"

	return appStyle.Render(board)
}

func RunTui() {
	addr := "127.0.0.1:3000"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	p := tea.NewProgram(initialModel(conn))

	go func() {
		newReader := bufio.NewReader(conn)
		for {
			data := refactorHandleMessage(newReader)
			// log.Println(data)
			p.Send(data)
			// p.Send(refactorHandleMessage(conn))
		}
	}()
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, theres been an error: %v", err)
		os.Exit(1)
	}
}
