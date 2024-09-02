package nfo

import "encoding/xml"

type nfoMovie struct {
	XMLName  xml.Name     `xml:"movie"`
	Tags     []string     `xml:"tag"`
	Genres   []string     `xml:"genre"`
	FileInfo *nfoFileInfo `xml:"fileinfo"`
}

type nfoEpisode struct {
	XMLName  xml.Name     `xml:"episodedetails"`
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
