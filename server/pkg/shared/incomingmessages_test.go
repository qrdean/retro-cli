package shared

import "testing"

func TestIncomingMessages(t *testing.T) {
	t.Run("marshal and unmarshal Add Sticky", func(t *testing.T) {
		var bytes [255]byte
		stringThing := []byte("Some Message")
		if len(stringThing) < 255 {
			copy(bytes[:len(stringThing)], stringThing)
		} else {
			t.Fatal("string isnt short")
		}

		addSticky := AddSticky{
			PosterId:      1,
			TopicId:       1,
			StickyMessage: bytes,
		}

		var addStickyBytes AddStickyBytes
		addStickyBytes = addSticky.MarshalBinary()
		newAddSticky := addStickyBytes.UnmarshalBinary()
		assert(t, newAddSticky.PosterId, addSticky.PosterId)
		assert(t, newAddSticky.TopicId, addSticky.TopicId)
		assert(t, newAddSticky.StickyMessage, addSticky.StickyMessage)
	})

	t.Run("marshal and unmarshal Vote Sticky", func(t *testing.T) {
		voteSticky := VoteSticky{
			StickyId: 1,
		}

		var voteStickyBytes VoteBytes
		voteStickyBytes = voteSticky.MarshalBinary()
		newAddSticky := voteStickyBytes.UnmarshalBinary()
		assert(t, newAddSticky.StickyId, voteSticky.StickyId)
	})

	t.Run("marshal and unmarshal Quit ", func(t *testing.T) {
		quit := Quit{
			ConnectionId: 1,
		}

		var quitBytes QuitBytes
		quitBytes = quit.MarshalBinary()
		newAddSticky := quitBytes.UnmarshalBinary()
		assert(t, newAddSticky.ConnectionId, quit.ConnectionId)
	})
}

func assert(t *testing.T, got, want interface{}) {
	if got != want {
		t.Fatalf("%v does not equal %v", got, want)
	}
}
