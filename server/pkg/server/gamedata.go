package server

import (
	"errors"
	"fmt"
	"pkg/shared"
)

type Board struct {
	PointToStickyId uint32
	StickyIdCounter uint32
	Topics          []Topic
}

type Topic struct {
	Id       uint32
	Header   string
	Stickies []Sticky
}

type Sticky struct {
	Id            uint32
	PosterId      uint32
	Votes         uint32
	StickyMessage string
}

func NewBoard(topics []Topic) Board {
	return Board{
		StickyIdCounter: 1,
		Topics:          topics,
	}
}

func NewTopic(id uint32, header string) Topic {
	return Topic{
		Id:     id,
		Header: header,
	}
}

func NewSticky(id, posterId, votes uint32, msg string) Sticky {
	return Sticky{
		Id:            id,
		PosterId:      posterId,
		Votes:         votes,
		StickyMessage: msg,
	}
}

func (b Board) AddNewTopic(header string) Board {
	id := len(b.Topics) + 1
	topic := NewTopic(uint32(id), header)
	b.Topics = append(b.Topics, topic)
	return b
}

func (b Board) FindSticky(stickyId uint32) (Sticky, int, int, error) {
	for boardIdx, topic := range b.Topics {
		sticky, stickyIndex, err := topic.findSticky(stickyId)
		if err != nil {
			continue
		}
		return sticky, stickyIndex, boardIdx, nil
	}
	return Sticky{}, -1, -1, errors.New("sticky not found")
}

func (b Board) PointToSticky(stickyId uint32) bool {
	found := false
	for _, topic := range b.Topics {
		_, _, err := topic.findSticky(stickyId)
		if err != nil {
			continue
		}
		found = true
	}
	return found
}

func (b Board) ToBoardMessages() ([]shared.Topic, []shared.Sticky, error) {
	var topics []shared.Topic
	var stickies []shared.Sticky
	for _, topic := range b.Topics {
		for _, sticky := range topic.Stickies {
			stickyMessage, err := sticky.toStickyMessage(topic.Id)
			if err != nil {
				return topics, stickies, err
			}
			stickies = append(stickies, stickyMessage)
		}
		topicMessage, err := topic.toTopicMessage()
		if err != nil {
			return topics, stickies, err
		}

		topics = append(topics, topicMessage)
	}

	return topics, stickies, nil
}

func (b Board) FindTopic(topicId uint32) (Topic, int, error) {
	for idx, topic := range b.Topics {
		if topic.Id == topicId {
			return topic, idx, nil
		}
	}

	return Topic{}, -1, errors.New("no topic found")
}

func (t Topic) AddNewSticky(newSticky Sticky) Topic {
	// id := len(t.Stickies) + 1
	// stickyPointer := &newSticky
	// stickyPointer.Id = uint32(id)
	t.Stickies = append(t.Stickies, newSticky)
	return t
}

func (t Topic) findSticky(stickyId uint32) (Sticky, int, error) {
	for idx, sticky := range t.Stickies {
		if sticky.Id == stickyId {
			return sticky, idx, nil
		}
	}

	return Sticky{}, -1, errors.New("no sticky found")
}

func (t Topic) toTopicMessage() (shared.Topic, error) {
	topicMessage, err := shared.NewTopic(t.Id, t.Header)
	if err != nil {
		return shared.Topic{}, err
	}
	return topicMessage, nil
}

func (s Sticky) VoteForSticky() Sticky {
	s.Votes++
	return s
}

func (s Sticky) toStickyMessage(topicId uint32) (shared.Sticky, error) {
	stickyMessage, err := shared.NewSticky(s.Id, s.PosterId, topicId, s.Votes, s.StickyMessage)
	if err != nil {
		return shared.Sticky{}, err
	}
	return stickyMessage, nil
}

func createEmptyTopicBoardStateState() Board {
	i := 0
	topics := []Topic{}
	for i < 3 {
		headerName := fmt.Sprintf("Topic %v", i)
		topics = append(topics, NewTopic(uint32(i), headerName))
		i++
	}
	return NewBoard(topics)
}

func createStickyTopicBoardStateState() Board {
	i := 0
	topics := []Topic{}
	for i < 3 {
		headerName := fmt.Sprintf("Topic %v", i)
		topics = append(topics, NewTopic(uint32(i), headerName))
		i++
	}
	stickyIds := 1
	for idx, topic := range topics {
		for numb := 0; numb < 100; numb++ {
			stickyMsg := fmt.Sprintf("Sticky %v", numb)
			sticky := NewSticky(uint32(stickyIds), 1, 0, stickyMsg)
			topic = topic.AddNewSticky(sticky)
			topics[idx] = topic
			stickyIds++
		}
	}
	board := NewBoard(topics)
	board.StickyIdCounter = uint32(stickyIds)
	return board
}
