package manager

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"strings"
)

type Video struct {
	Title       string `json:"title"`
	UploadDate  string `json:"upload_date"` // Format: "YYYYMMDD"
	Description string `json:"description"`
	WebpageURL  string `json:"webpage_url"`
	Thumbnail   string `json:"thumbnail"`
	Uploader    string `json:"uploader"`
	ViewCount   int    `json:"view_count,omitempty"`
	Duration    int    `json:"duration,omitempty"`
}

func FetchChannelVideos(channel string) ([]Video, error) {
	channelURL := channel
	if !strings.HasPrefix(channel, "http") {
		channelURL = "https://www.youtube.com/channel/" + channel
	}
	cmd := exec.Command("yt-dlp", channelURL, "--skip-download", "--dump-json", "--playlist-end", "20")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(stdout)
	var videos []Video
	for scanner.Scan() {
		line := scanner.Text()
		var video Video
		if err := json.Unmarshal([]byte(line), &video); err != nil {
			continue
		}
		videos = append(videos, video)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return videos, nil
}
