package nfo

import "regexp"

type fileContext int

const (
	Movie fileContext = iota
	Episode
)

var episodeRegex = regexp.MustCompile(`(?i)S(\d{2})E(\d{2})`)

func detectFileContext(filename string) fileContext {
	if episodeRegex.MatchString(filename) {
		return Episode
	}
	return Movie
}
