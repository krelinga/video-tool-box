package main

import "encoding/json"

type mkvmergeJson struct {
    Tracks []*jsonTrack `json:"tracks"`
}

type jsonTrack struct {
    Codec string `json:"codec"`
    Id int `json:"id"`
    Properties *jsonProperties `json:"properties"`
    Type string `json:"type"`
}

type jsonProperties struct {
    AudioChannels int `json:"audio_channels"`
    CodecId string `json:"codec_id"`
    DefaultTrack bool `json:"default_track"`
    EnabledTrack bool `json:"enabled_track"`
    ForcedTrack bool `json:"forced_track"`
    Language string `json:"language"`
    Number int `json:"number"`
    TrackName string `json:"track_name"`
    Uid int `json:"uid"`
}

func newMkvmergeJson(in []byte) (*mkvmergeJson, error) {
    var out mkvmergeJson
    if err := json.Unmarshal(in, &out); err != nil {
        return nil, err
    }
    return &out, nil
}
