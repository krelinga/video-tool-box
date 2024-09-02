package nfo

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type nfoMovie struct {
	XMLName  xml.Name     `xml:"movie"`
	Tags     []string     `xml:"tag"`
	Genres   []string     `xml:"genre"`
	FileInfo *nfoFileInfo `xml:"fileinfo"`
}

type nfoFileInfo struct {
	StreamDetails *nfoStreamDetails `xml:"streamdetails"`
}

type nfoStreamDetails struct {
	Video *nfoVideo `xml:"video"`
}

type nfoVideo struct {
	Width  int `xml:"width"`
	Height int `xml:"height"`
}

func readFromFile(filename string) (*nfoMovie, error) {
	// Open the XML file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new XML decoder
	decoder := xml.NewDecoder(file)

	// Parse the XML content
	entry := &nfoMovie{}
	err = decoder.Decode(entry)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("no movie data found")
		}
		return nil, err
	}

	return entry, nil
}

type Content struct {
	Tags   []string
	Genres []string
	Width  int
	Height int
}

func Parse(filename string) (*Content, error) {
	movie, err := readFromFile(filename)
	if err != nil {
		return nil, err
	}

	if movie.FileInfo == nil {
		return nil, fmt.Errorf("no file info found")
	}
	fileInfo := movie.FileInfo
	if fileInfo.StreamDetails == nil {
		return nil, fmt.Errorf("no stream details found")
	}
	streamDetails := fileInfo.StreamDetails
	if streamDetails.Video == nil {
		return nil, fmt.Errorf("no video stream details found")
	}
	video := streamDetails.Video
	if video.Width == 0 || video.Height == 0 {
		return nil, fmt.Errorf("invalid video resolution")
	}

	return &Content{
		Tags:   movie.Tags,
		Genres: movie.Genres,
		Width:  video.Width,
		Height: video.Height,
	}, nil
}
