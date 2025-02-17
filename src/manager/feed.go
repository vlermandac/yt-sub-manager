package manager

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
	"bufio"
	"encoding/json"
)

// Video represents a minimal set of metadata for a YouTube video.
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

// FetchFeedVideos retrieves video metadata from any given YouTube feed URL (for example,
// the home recommended feed or the subscriptions feed). It uses yt-dlp to dump JSON
// metadata for the last 20 videos.
func FetchFeedVideos(feedURL string) ([]Video, error) {
	// Assumes yt-dlp has cookies configured to access authenticated feeds.
	cmd := exec.Command("yt-dlp", feedURL, "--skip-download", "--dump-json", "--playlist-end", "20")
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

// GetFeedForCategory retrieves a combined feed of videos for the specified category.
// It supports three cases:
//   1. If category is "home", it fetches the authenticated home (recommended) feed.
//   2. If category is "subscriptions", it fetches the authenticated subscriptions feed.
//   3. Otherwise, it treats the category as a group of channels (loaded from subscriptions.json)
//      and fetches the latest videos from each channel concurrently.
func GetFeedForCategory(category string) ([]Video, error) {
	// Special case: home feed
	if strings.ToLower(category) == "home" {
		return FetchFeedVideos("https://www.youtube.com/feed/recommended")
	}
	// Special case: subscriptions feed
	if strings.ToLower(category) == "subscriptions" {
		return FetchFeedVideos("https://www.youtube.com/feed/subscriptions")
	}

	// For any other category, look up the channels stored in our persistent subscriptions.
	channels, err := GetChannels(category)
	if err != nil {
		return nil, err
	}
	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels in category '%s'", category)
	}

	var wg sync.WaitGroup
	videoCh := make(chan []Video, len(channels))
	errCh := make(chan error, len(channels))

	for _, ch := range channels {
		wg.Add(1)
		go func(ch string) {
			defer wg.Done()
			// Fetch videos for this channel.
			videos, err := FetchChannelVideos(ch)
			if err != nil {
				errCh <- fmt.Errorf("error fetching channel '%s': %v", ch, err)
				return
			}
			videoCh <- videos
		}(ch)
	}

	wg.Wait()
	close(videoCh)
	close(errCh)

	if len(errCh) > 0 {
		// Return the first error encountered.
		return nil, <-errCh
	}

	var allVideos []Video
	for vids := range videoCh {
		allVideos = append(allVideos, vids...)
	}

	// Sort videos by upload date (most recent first).
	sort.Slice(allVideos, func(i, j int) bool {
		t1, err1 := time.Parse("20060102", allVideos[i].UploadDate)
		t2, err2 := time.Parse("20060102", allVideos[j].UploadDate)
		if err1 == nil && err2 == nil {
			return t1.After(t2)
		}
		return allVideos[i].UploadDate > allVideos[j].UploadDate
	})

	// Limit to the last 20 videos.
	if len(allVideos) > 20 {
		allVideos = allVideos[:20]
	}

	return allVideos, nil
}
