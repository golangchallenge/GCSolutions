package instagram

import "errors"

// The Fetcher interface fetches media objects.
type Fetcher interface {
	// Fetch returns two channels. The first channel receives media
	// objects. The second channel is done. The caller must close(done) to
	// stop pulling images.
	Fetch() (chan *Media, chan struct{})
}

type tagFetcher struct {
	client Client
	tag    string
}

// NewTagFetcher gives you a Fetcher that pulls images for a tag.
func NewTagFetcher(c Client, t string) Fetcher {
	return &tagFetcher{c, t}
}

func (f tagFetcher) Fetch() (chan *Media, chan struct{}) {
	ch := make(chan *Media)
	done := make(chan struct{})
	var maxID string
	go func() {
		for {
			res, err := f.client.Tagged(f.tag, maxID)
			if err != nil || res == nil {
				return
			}
			for _, m := range res.Media {
				err := func(m Media) error {
					select {
					case ch <- &m:
					case <-done:
						return errors.New("abort")
					}
					return nil
				}(m)
				if err != nil {
					return
				}
			}
			maxID = res.MaxTagID
		}
	}()
	return ch, done
}
