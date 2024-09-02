package nfo

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type nfoRoot interface {
	nfoMovie | nfoEpisode | nfoShow
}

func readNfoFile[rootType nfoRoot](filename string, out *rootType) error {
	// Open the XML file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new XML decoder
	decoder := xml.NewDecoder(file)

	// Parse the XML content
	err = decoder.Decode(out)
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("no NFO data found")
		}
		return err
	}

	return nil
}

type Content struct {
	Tags   []string
	Genres []string
	Width  int
	Height int
}

func Parse(filename string) (*Content, error) {
	var fileInfo *nfoFileInfo
	var tags, genres []string
	switch detectFileContext(filename) {
	case Movie:
		var movie nfoMovie
		err := readNfoFile(filename, &movie)
		if err != nil {
			return nil, err
		}
		fileInfo = movie.FileInfo
		tags = movie.Tags
		genres = movie.Genres
	case Episode:
		var episode nfoEpisode
		err := readNfoFile[nfoEpisode](filename, &episode)
		if err != nil {
			return nil, err
		}
		fileInfo = episode.FileInfo

		showNfoPath, err := showNfoPath(filename)
		if err != nil {
			return nil, err
		}
		var show nfoShow
		if err := readNfoFile(showNfoPath, &show); err != nil {
			return nil, err
		}
		tags = show.Tags
		genres = show.Genres
	}

	if fileInfo == nil {
		return nil, fmt.Errorf("no file info found")
	}
	streamDetails := fileInfo.StreamDetails
	if streamDetails == nil {
		return nil, fmt.Errorf("no stream details found")
	}
	video := streamDetails.Video
	if video == nil {
		return nil, fmt.Errorf("no video stream details found")
	}
	if video.Width == 0 || video.Height == 0 {
		return nil, fmt.Errorf("invalid video resolution")
	}

	return &Content{
		Tags:   tags,
		Genres: genres,
		Width:  video.Width,
		Height: video.Height,
	}, nil
}
