package nfo

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename string
		expected *Content
		errMsg   string
	}{
		{
			filename: "../testdata/nfo/movies/Beavis and Butt-Head Do America (1996).nfo",
			expected: &Content{
				Tags:   []string{"hotel", "sperm", "washington dc, usa", "casino", "sun", "road trip", "las vegas", "adult animation", "based on tv series"},
				Genres: []string{"Animation", "Comedy"},
				Width:  720,
				Height: 480,
			},
			errMsg: "",
		},
		{
			filename: "../testdata/nfo/movies/The Void (2016).nfo",
			expected: &Content{
				Tags:   []string{"nurse", "mutation", "mutant", "morgue", "murder", "hospital", "another dimension", "doctor", "pregnant woman", "cosmic horror", "ax"},
				Genres: []string{"Mystery", "Horror", "Science Fiction"},
				Width:  1920,
				Height: 1080,
			},
			errMsg: "",
		},
		{
			filename: "../testdata/nfo/movies/They Live (1988).nfo",
			expected: &Content{
				Tags:   []string{"dystopia", "villainess", "alien", "social commentary", "conspiracy", "los angeles, california", "alien invasion", "sunglasses", "glasses", "brawl", "subliminal message", "horror"},
				Genres: []string{"Science Fiction", "Action"},
				Width:  720,
				Height: 480,
			},
			errMsg: "",
		},
		{
			filename: "../testdata/nfo/movies/errors/no_movie.nfo",
			expected: nil,
			errMsg:   "no NFO data found",
		},
		{
			filename: "../testdata/nfo/movies/errors/no_fileinfo.nfo",
			expected: nil,
			errMsg:   "no file info found",
		},
		{
			filename: "../testdata/nfo/movies/errors/no_streamdetails.nfo",
			expected: nil,
			errMsg:   "no stream details found",
		},
		{
			filename: "../testdata/nfo/movies/errors/no_video.nfo",
			expected: nil,
			errMsg:   "no video stream details found",
		},
		{
			filename: "../testdata/nfo/movies/errors/no_height.nfo",
			expected: nil,
			errMsg:   "invalid video resolution",
		},
		{
			filename: "../testdata/nfo/movies/errors/no_width.nfo",
			expected: nil,
			errMsg:   "invalid video resolution",
		},
		{
			filename: "../testdata/nfo/shows/Band of Brothers (2001)/Season 1/Band of Brothers - S01E01 - Currahee.nfo",
			expected: &Content{
				Genres: []string{"Mini-Series", "Drama", "Adventure", "Action", "History", "War"},
				Width:  720,
				Height: 480,
			},
			errMsg: "",
		},
		{
			filename: "../testdata/nfo/shows/Cowboy Bebop (1998)/Season 1/Cowboy Bebop - S01E01 - Asteroid Blues.nfo",
			expected: &Content{
				Genres: []string{"Science Fiction", "Drama", "Comedy", "Animation", "Adventure", "Action", "Western", "Anime"},
				Width:  720,
				Height: 480,
			},
			errMsg: "",
		},
		{
			filename: "../testdata/nfo/shows/The Terror (2018)/Season 1/The Terror - S01E01 - Go for Broke.nfo",
			expected: &Content{
				Genres: []string{"Horror", "Drama", "Adventure", "Suspense", "Thriller", "History"},
				Width:  1920,
				Height: 1080,
			},
			errMsg: "",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.filename, func(t *testing.T) {
			t.Parallel()
			nfoFile, err := os.Open(test.filename)
			if err != nil {
				t.Fatalf("Failed to open file %q: %v", test.filename, err)
			}
			content, err := Parse(test.filename, nfoFile)
			if err != nil {
				if test.errMsg == "" {
					t.Errorf("Unexpected error: %v", err)
				} else if err.Error() != test.errMsg {
					t.Errorf("Expected error message: %q, got: %q", test.errMsg, err.Error())
				}
				return
			}

			if test.errMsg != "" {
				t.Errorf("Expected error message: %q, got: nil", test.errMsg)
				return
			}

			if !equalContent(content, test.expected) {
				t.Errorf("Expected content: %v, got: %v", test.expected, content)
			}
		})
	}
}

func equalContent(c1, c2 *Content) bool {
	if c1 == nil && c2 == nil {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	if len(c1.Tags) != len(c2.Tags) {
		return false
	}
	for i := range c1.Tags {
		if c1.Tags[i] != c2.Tags[i] {
			return false
		}
	}
	if len(c1.Genres) != len(c2.Genres) {
		return false
	}
	for i := range c1.Genres {
		if c1.Genres[i] != c2.Genres[i] {
			return false
		}
	}
	return c1.Width == c2.Width && c1.Height == c2.Height
}