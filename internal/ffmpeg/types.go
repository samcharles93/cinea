package ffmpeg

import "time"

// MediaMetadata stores the extracted metadata
type MediaMetadata struct {
	Filename         string
	FormatName       string
	FormatLongName   string
	Duration         time.Duration
	Size             int64
	BitRate          int
	ProbeScore       int
	Container        string
	Codec            string
	ResolutionWidth  int
	ResolutionHeight int
	FrameRate        string
	AudioTracks      []AudioTrackMetadata
	VideoTracks      []VideoTrackMetadata
	SubtitleTracks   []SubtitleTrackMetadata
	Tags             map[string]string
}

// AudioTrackMetadata stores information about a single audio track
type AudioTrackMetadata struct {
	Index       int
	Codec       string
	Channels    int
	SampleRate  string
	Language    string
	BitRate     int
	Tags        map[string]string
	Disposition map[string]int
}

// VideoTrackMetadata stores information about a single video track
type VideoTrackMetadata struct {
	Index              int
	CodecName          string
	CodecLongName      string
	Profile            string
	Width              int
	Height             int
	CodedWidth         int
	CodedHeight        int
	HasBFrames         int
	SampleAspectRatio  string
	DisplayAspectRatio string
	PixFmt             string
	Level              int
	ColorRange         string
	ColorSpace         string
	ColorTransfer      string
	ColorPrimaries     string
	ChromaLocation     string
	FieldOrder         string
	Refs               int
	FrameRate          string
	Tags               map[string]string
	Disposition        map[string]int
	SideDataList       []SideData
}

type SideData struct {
	SideDataType string `json:"side_data_type"`

	// Dolby Vision data
	DVVersionMajor *int   `json:"dv_version_major,omitempty"`
	DVVersionMinor *int   `json:"dv_version_minor,omitempty"`
	DVProfile      *int   `json:"dv_profile,omitempty"`
	DVLevel        *int   `json:"dv_level,omitempty"`
	DVBLSignal     *int   `json:"dv_bl_signal_compatibility_id,omitempty"`
	DVMD           string `json:"dv_md_compression,omitempty"`

	RPU *int `json:"rpu_present_flag,omitempty"`
	EL  *int `json:"el_present_flag,omitempty"`
	BL  *int `json:"bl_present_flag,omitempty"`
}

// SubtitleTrackMetadata stores info about a subtitle track
type SubtitleTrackMetadata struct {
	Index         int
	CodecName     string
	CodecLongName string
	CodecType     string
	Tags          map[string]string
	Disposition   map[string]int
}
