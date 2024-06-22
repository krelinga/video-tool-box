package hb

import "errors"

type Flags []string

var profiles = map[string]Flags{
    "mkv_h265_1080p30": {
        "-Z", "Matroska/H.265 MKV 1080p30",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2160p60_surround": {
        "-Z", "General/Super HQ 2160p60 4K HEVC Surround",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2160p60_no_audio": {
        "-Z", "General/Super HQ 2160p60 4K HEVC Surround",
        "--audio", "none",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_1080p30_no_sound": {
        "-Z", "Matroska/H.265 MKV 1080p30",
        "--audio", "none",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2160p60_fast": {
        "-Z", "General/Fast 2160p60 4K HEVC",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2160p60_very_fast": {
        "-Z", "General/Very Fast 2160p60 4K HEVC",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2060p60_4k": {
        "-Z", "Matroska/H.265 MKV 2160p60 4K",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2060p60_4k_28.0": {
        "-Z", "Matroska/H.265 MKV 2160p60 4K",
        "-q", "28.0",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2060p60_4k_18.0": {
        "-Z", "Matroska/H.265 MKV 2160p60 4K",
        "-q", "18.0",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2060p60_4k_19.0": {
        "-Z", "Matroska/H.265 MKV 2160p60 4K",
        "-q", "19.0",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
    "mkv_h265_2060p60_4k_20.0": {
        "-Z", "Matroska/H.265 MKV 2160p60 4K",
        "-q", "20.0",
        "--all-audio",
        "--non-anamorphic",
        "--all-subtitles",
        "--subtitle-burned=none",
    },
}

var ProfileNotFoundErr = errors.New("Profile not found")

func GetFlags(profile, inPath, outPath string) (Flags, error) {
    profileFlags, found := profiles[profile]
    if !found {
        return nil, ProfileNotFoundErr
    }
    standardFlags := []string{
        "-i", inPath,
        "-o", outPath,
        "--json",
    }
    flags := append(standardFlags, profileFlags...)
    return flags, nil
}
