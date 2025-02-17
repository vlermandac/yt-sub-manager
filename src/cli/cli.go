package cli

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/yt-sub-manager/src/manager"
)

// RunCLI dispatches CLI commands based on arguments.
func RunCLI(args []string) {
	// Ensure yt-dlp is installed.
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		fmt.Fprintln(os.Stderr, "Error: yt-dlp is not installed or not found in PATH.")
		os.Exit(1)
	}

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "category":
		handleCategory(args[1:])
	case "channel":
		handleChannel(args[1:])
	case "feed":
		handleFeed(args[1:])
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	usage := `
CLI Usage:
  category create <name>
  category delete <name>
  channel add <category> <channel>
  channel remove <category> <channel>
  feed <category>
`
	fmt.Println(usage)
}

func handleCategory(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Not enough arguments for category command.")
		printUsage()
		os.Exit(1)
	}
	action, categoryName := args[0], args[1]
	switch action {
	case "create":
		if err := manager.CreateCategory(categoryName); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("Category '%s' created successfully.\n", categoryName)
	case "delete":
		if err := manager.DeleteCategory(categoryName); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("Category '%s' deleted successfully.\n", categoryName)
	default:
		fmt.Fprintf(os.Stderr, "Unknown category action: %s\n", action)
		printUsage()
		os.Exit(1)
	}
}

func handleChannel(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "Error: Not enough arguments for channel command.")
		printUsage()
		os.Exit(1)
	}
	action, categoryName, channel := args[0], args[1], args[2]
	switch action {
	case "add":
		if err := manager.AddChannelToCategory(categoryName, channel); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("Channel '%s' added to category '%s'.\n", channel, categoryName)
	case "remove":
		if err := manager.RemoveChannelFromCategory(categoryName, channel); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("Channel '%s' removed from category '%s'.\n", channel, categoryName)
	default:
		fmt.Fprintf(os.Stderr, "Unknown channel action: %s\n", action)
		printUsage()
		os.Exit(1)
	}
}

func handleFeed(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Category name required for feed command.")
		printUsage()
		os.Exit(1)
	}
	categoryName := args[0]
	feed, err := manager.GetFeedForCategory(categoryName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if len(feed) == 0 {
		fmt.Println("No videos found in the channels for this category.")
		return
	}
	fmt.Printf("Feed for category '%s' (last %d videos):\n", categoryName, len(feed))
	fmt.Println("------------------------------------------------------------")
	for _, video := range feed {
		fmt.Printf("Title       : %s\n", video.Title)
		fmt.Printf("URL         : %s\n", video.WebpageURL)
		fmt.Printf("Thumbnail   : %s\n", video.Thumbnail)
		pubDate := video.UploadDate
		if t, err := time.Parse("20060102", video.UploadDate); err == nil {
			pubDate = t.Format("2006-01-02")
		}
		fmt.Printf("Published   : %s\n", pubDate)
		fmt.Printf("Channel     : %s\n", video.Uploader)
		if video.ViewCount != 0 {
			fmt.Printf("View Count  : %d\n", video.ViewCount)
		}
		if video.Duration != 0 {
			dur := time.Duration(video.Duration) * time.Second
			fmt.Printf("Duration    : %s\n", dur)
		}
		desc := video.Description
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		fmt.Printf("Description : %s\n", desc)
		fmt.Println("------------------------------------------------------------")
	}
}
