package show

import (
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

var seasonRE = regexp.MustCompile(`Season \d+`)

func FindEpisodes(dir string) ([]string, error) {
    // Find the season directories.
    filesInDir, err := os.ReadDir(dir)
    if err != nil {
        return nil, err
    }
    seasons := []string{}
    for _, file := range filesInDir {
        if seasonRE.MatchString(file.Name()) {
            seasons = append(seasons, file.Name())
        }
    }

    episodes := []string{}
    // Find the episodes in each season directory
    for _, season := range seasons {
        seasonFiles, err := os.ReadDir(filepath.Join(dir, season))
        if err != nil {
            return nil, err
        }
        for _, file := range seasonFiles {
            if filepath.Ext(file.Name()) == ".mkv" {
                episodes = append(episodes, filepath.Join(dir, season, file.Name()))
            }
        }
    }
    return episodes, nil
}

func MapPaths(inPaths []string, inDir, outDir string) map[string]string {
    out := make(map[string]string)
    for _, p := range inPaths {
        child, found := strings.CutPrefix(p, inDir)
        if !found {
            panic(p)
        }
        out[p] = outDir + child
    }
    return out
}
