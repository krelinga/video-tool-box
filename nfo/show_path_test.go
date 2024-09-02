package nfo

import (
	"errors"
	"testing"
)

func TestShowNfoPath(t *testing.T) {
	tests := []struct {
		episodeNfoPath string
		expectedPath   string
		expectedError  error
	}{
		{
			episodeNfoPath: "/nas/media/Shows/Band of Brothers (2001)/Season 1/Band of Brothers - S01E01 - Currahee.nfo",
			expectedPath:   "/nas/media/Shows/Band of Brothers (2001)/tvshow.nfo",
			expectedError:  nil,
		},
		{
			episodeNfoPath: "/nas/media/Shows/Cowboy Bebop (1998)/Season 1/Cowboy Bebop - S04E01 - Asteroid Blues.nfo",
			expectedPath:   "/nas/media/Shows/Cowboy Bebop (1998)/tvshow.nfo",
			expectedError:  nil,
		},
		{
			episodeNfoPath: "/nas/media/Shows/The Terror (2018)/Season 1/The Terror - S01E78 - Go for Broke.nfo",
			expectedPath:   "/nas/media/Shows/The Terror (2018)/tvshow.nfo",
			expectedError:  nil,
		},
		{
			episodeNfoPath: "/nas/media/Movies/Beavis and Butt-Head Do America (1996)/Beavis and Butt-Head Do America (1996).nfo",
			expectedPath:   "",
			expectedError:  errors.New("file is not an episode NFO"),
		},
		{
			episodeNfoPath: "The Terror - S01E78 - Go for Broke.nfo",
			expectedPath:   "",
			expectedError:  errors.New("invalid path"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.episodeNfoPath, func(t *testing.T) {
			t.Parallel()
			path, err := showNfoPath(test.episodeNfoPath)
			if path != test.expectedPath {
				t.Errorf("Expected path: %v, got: %v", test.expectedPath, path)
			}
			if (err == nil && test.expectedError != nil) || (err != nil && test.expectedError == nil) || (err != nil && test.expectedError != nil && err.Error() != test.expectedError.Error()) {
				t.Errorf("Expected error: %v, got: %v", test.expectedError, err)
			}
		})
	}
}
