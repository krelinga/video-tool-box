package nfo

import "testing"

func TestDetectFileContext(t *testing.T) {
	tests := []struct {
		filename     string
		expectedType fileContext
	}{
		{
			filename:     "/nas/media/Movies/Beavis and Butt-Head Do America (1996)/Beavis and Butt-Head Do America (1996).nfo",
			expectedType: Movie,
		},
		{
			filename:     "/nas/media/Movies/The Void (2016)/The Void (2016).nfo",
			expectedType: Movie,
		},
		{
			filename:     "/nas/media/Movies/They Live (1988)/They Live (1988).nfo",
			expectedType: Movie,
		},
		{
			filename:     "/nas/media/Shows/Band of Brothers (2001)/Season 1/Band of Brothers - S01E01 - Currahee.nfo",
			expectedType: Episode,
		},
		{
			filename:     "/nas/media/Shows/Cowboy Bebop (1998)/Season 1/Cowboy Bebop - S04E01 - Asteroid Blues.nfo",
			expectedType: Episode,
		},
		{
			filename:     "/nas/media/Shows/The Terror (2018)/Season 1/The Terror - S01E78 - Go for Broke.nfo",
			expectedType: Episode,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.filename, func(t *testing.T) {
			t.Parallel()
			fileType := detectFileContext(test.filename)
			if fileType != test.expectedType {
				t.Errorf("Expected file context: %v, got: %v", test.expectedType, fileType)
			}
		})
	}
}
