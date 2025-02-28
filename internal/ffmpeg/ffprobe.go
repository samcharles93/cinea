package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

// RunFFprobe executes an FFprobe command with the provided arguments
func (s *service) RunFFprobe(ctx context.Context, args []string) ([]byte, error) {
	if err := s.EnsureInstalled(); err != nil {
		return nil, fmt.Errorf("failed to ensure FFprobe is installed: %w", err)
	}

	s.appLogger.Debug().Strs("args", args).Msg("Running FFprobe command")
	cmd := exec.CommandContext(ctx, s.ffprobePath, args...)
	output, err := cmd.Output()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			s.appLogger.Warn().Err(err).Str("stderr", string(exitError.Stderr)).Msg("FFprobe command had a non-zero exit code")
			return output, fmt.Errorf("ffprobe command failed with stderr: %s, error: %w",
				string(exitError.Stderr), err)
		} else {
			s.appLogger.Error().Err(err).Msg("Failed to execute ffprobe command")
			return nil, fmt.Errorf("ffprobe command failed: %w", err)
		}
	}

	return output, nil
}

// ExtractMetadata extracts metadata from the given media file using ffprobe
func (s *service) ExtractMetadata(ctx context.Context, filePath string) (*MediaMetadata, error) {
	s.appLogger.Info().
		Str("filepath", filePath).
		Msg("Starting media file metadata extraction")
	start := time.Now()

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format", "-show_streams",
		"-i", filePath,
	}

	output, err := s.RunFFprobe(ctx, args)

	// Store original error to return later if needed
	var ffprobeError error
	if err != nil {
		ffprobeError = err
		// We continue anyway as we might have partial output
	}

	metadata, parseErr := s.parseFFprobeJSONOutput(output)
	if parseErr != nil {
		return nil, parseErr
	}

	metadata.Filename = filePath
	duration := time.Since(start)
	s.appLogger.Info().
		Str("filepath", filePath).
		Dur("duration", duration).
		Msg("Metadata extraction complete")

	// Return the metadata along with any error that might have occurred
	return metadata, ffprobeError
}

// parseFFprobeJSONOutput parses the JSON output from ffprobe
func (s *service) parseFFprobeJSONOutput(output []byte) (*MediaMetadata, error) {
	var ffprobeData struct {
		Format struct {
			Filename       string            `json:"filename"`
			FormatName     string            `json:"format_name"`
			FormatLongName string            `json:"format_long_name"`
			Duration       string            `json:"duration"`
			Size           string            `json:"size"`
			BitRate        string            `json:"bit_rate"`
			ProbeScore     int               `json:"probe_score"`
			Tags           map[string]string `json:"tags"`
		} `json:"format"`
		Streams []struct {
			Index         int    `json:"index"`
			CodecName     string `json:"codec_name"`
			CodecType     string `json:"codec_type"`
			CodecLongName string `json:"codec_long_name"`
			Profile       string `json:"profile"`

			Width              int    `json:"width"`
			Height             int    `json:"height"`
			CodedWidth         int    `json:"coded_width"`
			CodedHeight        int    `json:"coded_height"`
			HasBFrames         int    `json:"has_b_frames"`
			AvgFrameRate       string `json:"avg_frame_rate"`
			SampleAspectRatio  string `json:"sample_aspect_ratio"`
			DisplayAspectRatio string `json:"display_aspect_ratio"`
			PixFmt             string `json:"pix_fmt"`
			Level              int    `json:"level"`
			ColorRange         string `json:"color_range"`
			ColorSpace         string `json:"color_space"`
			ColorTransfer      string `json:"color_transfer"`
			ColorPrimaries     string `json:"color_primaries"`
			ChromaLocation     string `json:"chroma_location"`
			FieldOrder         string `json:"field_order"`
			Refs               int    `json:"refs"`
			TimeBase           string `json:"time_base"`
			StartPts           int64  `json:"start_pts"`
			StartTime          string `json:"start_time"`

			Channels   int    `json:"channels"`
			SampleRate string `json:"sample_rate"`
			BitRate    string `json:"bit_rate"`

			Disposition  map[string]int    `json:"disposition"`
			Tags         map[string]string `json:"tags"`
			SideDataList []SideData        `json:"side_data_list"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &ffprobeData); err != nil {
		s.appLogger.Error().
			Err(err).
			Str("output", string(output)).
			Msg("Failed to parse JSON output")
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	metadata := &MediaMetadata{
		FormatName:     ffprobeData.Format.FormatName,
		FormatLongName: ffprobeData.Format.FormatLongName,
		ProbeScore:     ffprobeData.Format.ProbeScore,
		Container:      ffprobeData.Format.FormatName,
		Tags:           ffprobeData.Format.Tags,
	}

	// Parse duration (handle errors)
	durationSec, err := strconv.ParseFloat(ffprobeData.Format.Duration, 64)
	if err != nil {
		s.appLogger.Warn().
			Err(err).
			Msg("Failed to parse duration")
	} else {
		metadata.Duration = time.Duration(durationSec * float64(time.Second))
	}

	// Parse size (handle errors)
	size, err := strconv.ParseInt(ffprobeData.Format.Size, 10, 64)
	if err != nil {
		s.appLogger.Warn().
			Err(err).
			Msg("Failed to parse size")
	}
	metadata.Size = size

	// Parse bit rate (handle errors)
	bitRate, err := strconv.Atoi(ffprobeData.Format.BitRate)
	if err != nil {
		s.appLogger.Warn().
			Err(err).
			Msg("Could not parse bitrate")
	}
	metadata.BitRate = bitRate

	for _, stream := range ffprobeData.Streams {
		switch stream.CodecType {
		case "video":
			videoTrack := VideoTrackMetadata{
				Index:              stream.Index,
				CodecName:          stream.CodecName,
				CodecLongName:      stream.CodecLongName,
				Profile:            stream.Profile,
				Width:              stream.Width,
				Height:             stream.Height,
				CodedWidth:         stream.CodedWidth,
				CodedHeight:        stream.CodedHeight,
				HasBFrames:         stream.HasBFrames,
				FrameRate:          stream.AvgFrameRate,
				SampleAspectRatio:  stream.SampleAspectRatio,
				DisplayAspectRatio: stream.DisplayAspectRatio,
				PixFmt:             stream.PixFmt,
				Level:              stream.Level,
				ColorRange:         stream.ColorRange,
				ColorSpace:         stream.ColorSpace,
				ColorTransfer:      stream.ColorTransfer,
				ColorPrimaries:     stream.ColorPrimaries,
				ChromaLocation:     stream.ChromaLocation,
				FieldOrder:         stream.FieldOrder,
				Refs:               stream.Refs,
				Tags:               stream.Tags,
				Disposition:        stream.Disposition,
				SideDataList:       stream.SideDataList,
			}

			// Set the video track codec, as this is consistent across all media
			metadata.Codec = stream.CodecName
			metadata.ResolutionWidth = stream.Width
			metadata.ResolutionHeight = stream.Height
			metadata.FrameRate = stream.AvgFrameRate

			metadata.VideoTracks = append(metadata.VideoTracks, videoTrack)
		case "audio":
			bitrate := 0
			if stream.BitRate != "" {
				bitrate, err = strconv.Atoi(stream.BitRate)
				if err != nil {
					s.appLogger.Warn().Err(err).Msg("Could not parse audio bit rate")
				}
			}
			audioTrack := AudioTrackMetadata{
				Index:       stream.Index,
				Codec:       stream.CodecName,
				Channels:    stream.Channels,
				SampleRate:  stream.SampleRate,
				BitRate:     bitrate,
				Tags:        stream.Tags,
				Disposition: stream.Disposition,
			}

			// Extract language tag if available
			if lang, ok := stream.Tags["language"]; ok {
				audioTrack.Language = lang
			}
			metadata.AudioTracks = append(metadata.AudioTracks, audioTrack)

		case "subtitle":
			subtitleTrack := SubtitleTrackMetadata{
				Index:         stream.Index,
				CodecName:     stream.CodecName,
				CodecLongName: stream.CodecLongName,
				CodecType:     stream.CodecType,
				Tags:          stream.Tags,
				Disposition:   stream.Disposition,
			}
			metadata.SubtitleTracks = append(metadata.SubtitleTracks, subtitleTrack)
		default:
			s.appLogger.Debug().Str("codec_type", stream.CodecType).Msg("Skipping unsupported stream type")
		}
	}

	return metadata, nil
}
