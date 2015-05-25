package instagram

import (
	"fmt"
	"testing"
)

type tagCall struct {
	Tag   string
	MaxID string
}

type fakeClient struct {
	mediaLists     []*MediaList
	mediaListIndex int
	taggedCalls    []tagCall
}

func (c *fakeClient) Popular() (*MediaList, error) {
	return nil, nil
}

func (c *fakeClient) Search(oat, lng string) (*MediaList, error) {
	return nil, nil
}

func (c *fakeClient) Tagged(tag string, maxID string) (*MediaList, error) {
	c.taggedCalls = append(c.taggedCalls, tagCall{
		tag,
		maxID,
	})
	if c.mediaListIndex < len(c.mediaLists) {
		list := c.mediaLists[c.mediaListIndex]
		c.mediaListIndex++
		return list, nil
	}
	return nil, fmt.Errorf("no more data")
}

func Test_tagFetcher_Fetch(t *testing.T) {
	c := &fakeClient{}
	f := tagFetcher{c, "cat"}

	list1 := &MediaList{
		Media: []Media{
			{Type: "i1"},
			{Type: "i2"},
		},
		Pagination: Pagination{
			NextURL:  "",
			MaxTagID: "list2",
		},
	}
	list2 := &MediaList{
		Media: []Media{
			{Type: "i3"},
			{Type: "i4"},
		},
		Pagination: Pagination{
			NextURL:  "",
			MaxTagID: "list3",
		},
	}

	c.mediaLists = []*MediaList{list1, list2}
	ch, done := f.Fetch()

	var m *Media
	var media []*Media

	go func() {
		for {
			select {
			case m = <-ch:
				media = append(media, m)
				if len(media) == 3 {
					close(done)
					return
				}
			}
		}
	}()
	<-done

	if got, want := len(media), 3; got != want {
		t.Errorf("got %d records, want %d", got, want)
	}
	if got, want := len(c.taggedCalls), 2; got != want {
		t.Errorf("got %d calls, want %d. %#v", got, want, c.taggedCalls)
	}
	if got, want := c.taggedCalls[0].MaxID, ""; got != want {
		t.Errorf("calls 0, want %s, got %s", got, want)
	}
	if got, want := c.taggedCalls[1].MaxID, "list2"; got != want {
		t.Errorf("calls 1, want %s, got %s", got, want)
	}
}
