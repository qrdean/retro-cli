package shared

import "testing"

func TestBoard(t *testing.T) {
	t.Run("marshal and unmarshal Sticky", func(t *testing.T) {
		var bytes [255]byte
		stringThing := []byte("Some Message")
		if len(stringThing) < 255 {
			copy(bytes[:len(stringThing)], stringThing)
		} else {
			t.Fatal("string isnt short")
		}

		sticky := Sticky{
			Id:            1,
			PosterId:      2,
			Votes:         6,
			StickyMessage: bytes,
		}

		data := sticky.MarshalBinary()

		someSticky := UnmarshalBinaryStick(data)
		t.Logf("id: %v, posterid: %v, votes: %v, sticky message: %v\n", someSticky.Id, someSticky.PosterId, someSticky.Votes, string(someSticky.StickyMessage[:]))
	})

	t.Run("marshal and unmarshal Topic", func(t *testing.T) {
		topicHeaderString := []byte("Topic Header")

		var newBytes [255]byte
		copy(newBytes[:len(topicHeaderString)], topicHeaderString)

		topic := Topic{
			Id:     1,
			Header: newBytes,
		}

		data := topic.MarshalBinary()

		someTopic := UnmarshalTopic(data)
		t.Logf("Id: %v, header: %v", someTopic.Id, string(someTopic.Header[:]))
	})

	t.Run("marshal and unmarshal Pointer", func(t *testing.T) {

		pointer := Pointer{
			PointerId:     1,
		}

		data := pointer.MarshalBinary()

		somePointer := UnmarshalPointer(data)
		t.Logf("Id: %v", somePointer.PointerId)
	})
}
