package nfo

import (
	"io"
	"os"
	"slices"

	"github.com/krelinga/go-lib/video/nfo"
)

type Content struct {
	Tags   []string
	Genres []string
	Width  int
	Height int
}

func Parse(filename string, reader io.Reader) (*Content, error) {
	content := &Content{}
	raw, err := nfo.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	switch detectFileContext(filename) {
	case Movie:
		movieNfo := raw.(*nfo.Movie)
		content.Tags = slices.Collect(movieNfo.Tags())
		content.Genres = slices.Collect(movieNfo.Genres())
		content.Width = movieNfo.Width()
		content.Height = movieNfo.Height()
	case Episode:
		episodeNfo := raw.(*nfo.Episode)
		content.Width = episodeNfo.Width()
		content.Height = episodeNfo.Height()

		showNfoPath, err := showNfoPath(filename)
		if err != nil {
			return nil, err
		}
		showNfoOpened, err := os.Open(showNfoPath)
		if err != nil {
			return nil, err
		}
		rawShow, err := nfo.ReadFrom(showNfoOpened)
		if err != nil {
			return nil, err
		}
		showNfo := rawShow.(*nfo.TvShow)
		content.Tags = slices.Collect(showNfo.Tags())
		content.Genres = slices.Collect(showNfo.Genres())
	}

	return content, nil
}
