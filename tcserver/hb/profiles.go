package hb

import "errors"

type flags []string

var profiles = map[string]flags{
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

var (
    ProfileNotFoundErr = errors.New("Profile not found")
    InvalidInPathErr = errors.New("Invalid inPath")
    InvalidOutPathErr = errors.New("Invalid outPath")
)

func getFlags(profile, inPath, outPath string) (flags, error) {
    profileFlags, found := profiles[profile]
    if !found {
        return nil, ProfileNotFoundErr
    }
    if len(inPath) == 0 {
        return nil, InvalidInPathErr
    }
    if len(outPath) == 0 {
        return nil, InvalidOutPathErr
    }
    standardFlags := []string{
        "-i", inPath,
        "-o", outPath,
        "--json",
    }
    combinedFlags := append(standardFlags, profileFlags...)
    return combinedFlags, nil
}
