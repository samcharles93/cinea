package scanner

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	videoExts := map[string]bool{
		".mp4":  true,
		".mkv":  true,
		".avi":  true,
		".mov":  true,
		".m4v":  true,
		".webm": true,
		".wmv":  true, // Added .wmv
		".flv":  true, // Added .flv
		".ts":   true, // Added .ts (Transport Stream - common for recordings)
	}
	return videoExts[ext]
}

func isLikelyTVFile(path string) bool {
	filename := filepath.Base(path)
	return strings.Contains(filename, "S0") || strings.Contains(filename, "E0") || strings.Contains(strings.ToLower(filename), "s0") || strings.Contains(strings.ToLower(filename), "e0")
}

func getPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func cleanTitle(title string) string {
	// Replace dots, underscores with spaces
	title = strings.NewReplacer(".", " ", "_", " ").Replace(title)

	// Remove multiple spaces
	title = strings.Join(strings.Fields(title), " ")

	return title
}

func extractMovieInfo(path string) mediaInfo {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	nameOnly := strings.TrimSuffix(filename, ext)

	// Improved regex to handle more variations:
	// - Optional spaces around the year
	// - Year in square brackets: [2023]
	// - Year at the end, or potentially with resolution after:  Movie.Name.2023.1080p
	re := regexp.MustCompile(`^(.*?)(?:\s*\((\d{4})\)|\s*\[(\d{4})\]|\.(\d{4})\.)`)
	matches := re.FindStringSubmatch(nameOnly)

	if len(matches) > 0 {
		title := strings.TrimSpace(matches[1])
		year := ""
		// Find the first non-empty year match (group 2, 3, or 4)
		for i := 2; i < len(matches); i++ {
			if matches[i] != "" {
				year = matches[i]
				break
			}
		}
		return mediaInfo{
			Title: title,
			Year:  year,
		}
	}

	return mediaInfo{
		Title: nameOnly,
	}
}

func extractTVShowInfo(path string) tvShowInfo {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	nameOnly := strings.TrimSuffix(filename, ext)

	// Common TV show patterns:
	// Show Name S01E01
	// Show.Name.1x01
	// Show_Name_101
	// Show Name - S01E01
	// Show Name - 01x01
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(.+?)[\. _-]+S(\d{1,2})E(\d{1,2})`),
		regexp.MustCompile(`(?i)^(.+?)[\. _-]+(\d{1,2})x(\d{1,2})`),
		regexp.MustCompile(`(?i)^(.+?)[\. _-]+(\d)(\d{2})`),
	}

	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(nameOnly); len(matches) == 4 {
			season, _ := strconv.Atoi(matches[2])
			episode, _ := strconv.Atoi(matches[3])
			return tvShowInfo{
				Title:   cleanTitle(matches[1]),
				Season:  season,
				Episode: episode,
			}
		}
	}

	return tvShowInfo{
		Title: cleanTitle(nameOnly),
	}
}
