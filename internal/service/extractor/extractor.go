package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/samcharles93/cinea/internal/ffmpeg"
	"github.com/samcharles93/cinea/internal/logger"
)

type Service interface {
	Extract(ctx context.Context, filePath string) (*ffmpeg.MediaMetadata, error)
	parseFFprobeJSONOutput(output []byte) (*ffmpeg.MediaMetadata, error)
}

type service struct {
	appLogger logger.Logger
	ffService ffmpeg.Service
}

// NewExtractor creates a new Extractor, ensuring FFProbePath is set.
func NewExtractor(appLogger logger.Logger, ffService ffmpeg.Service) Service {
	ffService.GetFFprobePath()

	return &service{
		appLogger: appLogger,
		ffService: ffService,
	}
}

// Extract extracts metadata from the given file.
func (s *service) Extract(ctx context.Context, filePath string) (*ffmpeg.MediaMetadata, error) {
	s.appLogger.Info().Str("filepath", filePath).Msg("Starting media file extraction")
	start := time.Now()

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format", "-show_streams",
		"-i", filePath,
	}
	output, err := s.ffService.RunFFprobe(ctx, args)
	if err != nil {
		s.appLogger.Error().Err(err).Msg("Failed to extract media metadata")
		return nil, fmt.Errorf("failed to extract media metadata: %w", err)
	}

	// ffprobe can produce a non-zero error, we need to handle those cases
	var ffprobeError error

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			s.appLogger.Warn().
				Err(err).
				Str("filepath", filePath).
				Str("stderr", string(exitError.Stderr)).
				Msg("FFProbe command had a non-zero exit code")
			ffprobeError = fmt.Errorf("ffprobe command failed with stderr: %s, error: %w", string(exitError.Stderr), err)
		} else {
			s.appLogger.Error().
				Err(err).
				Str("filepath", filePath).
				Msg("Failed to execute ffprobe command")
			return nil, fmt.Errorf("ffprobe command failed: %w", err)
		}
	}

	metadata, err := s.parseFFprobeJSONOutput(output)
	if err != nil {
		return nil, err
	}

	metadata.Filename = filePath
	duration := time.Since(start)
	s.appLogger.Info().
		Str("filepath", filePath).
		Dur("duration", duration).
		Msg("Metadata extraction complete")
	return metadata, ffprobeError
}

// parseFFprobeJSONOutput
func (s *service) parseFFprobeJSONOutput(output []byte) (*ffmpeg.MediaMetadata, error) {
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
			SideDataList []ffmpeg.SideData `json:"side_data_list"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &ffprobeData); err != nil {
		s.appLogger.Error().
			Err(err).
			Str("output", string(output)).
			Msg("Failed to parse JSON output")
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	metadata := &ffmpeg.MediaMetadata{
		FormatName:     ffprobeData.Format.FormatName,
		FormatLongName: ffprobeData.Format.FormatLongName,
		ProbeScore:     ffprobeData.Format.ProbeScore,
		Container:      ffprobeData.Format.FormatName,
		Tags:           ffprobeData.Format.Tags,
	}

	// Parse duration (handle errors)
	durationSec, err := strconv.ParseFloat(ffprobeData.Format.Duration, 64)
	if err != nil {
		s.appLogger.Warn().Err(err).Msg("Failed to parse duration")
	} else {
		metadata.Duration = time.Duration(durationSec * float64(time.Second))
	}

	// Parse size (handle errors)
	size, err := strconv.ParseInt(ffprobeData.Format.Size, 10, 64)
	if err != nil {
		s.appLogger.Warn().Err(err).Msg("Failed to parse size")
	}
	metadata.Size = size
	bitRate, err := strconv.Atoi(ffprobeData.Format.BitRate)
	if err != nil {
		s.appLogger.Warn().
			Err(err).
			Msg("could not parse bitrate")
	}
	metadata.BitRate = bitRate

	for _, stream := range ffprobeData.Streams {
		switch stream.CodecType {
		case "video":
			videoTrack := ffmpeg.VideoTrackMetadata{
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

			//set the video track codec, as this is consistent across all media
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
					s.appLogger.Warn().Err(err).Msg("could not parse audio bit rate")
				}
			}
			audioTrack := ffmpeg.AudioTrackMetadata{
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
			subtitleTrack := ffmpeg.SubtitleTrackMetadata{
				Index:         stream.Index,
				CodecName:     stream.CodecName,
				CodecLongName: stream.CodecLongName,
				CodecType:     stream.CodecType,
				Tags:          stream.Tags,
				Disposition:   stream.Disposition,
			}
			metadata.SubtitleTracks = append(metadata.SubtitleTracks, subtitleTrack)
		default:
			s.appLogger.Debug().
				Str("codec_type", stream.CodecType).
				Msg("Skipping unsupported stream type")
		}
	}

	return metadata, nil
}
