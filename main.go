package main

import (
	"fmt"
	"os"

	"github.com/yt-sub-manager/src/cli"
	"github.com/yt-sub-manager/src/tui"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	mode := os.Args[1]
	switch mode {
	case "tui":
		tui.RunTUI()
	case "cli":
		cli.RunCLI(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", mode)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	usage := `
Usage:
  submanager cli <command> [arguments]
  submanager tui
  submanager help

CLI Commands:
  category create <name>            Create a new category.
  category delete <name>            Delete an existing category.
  channel add <category> <channel>  Add a YouTube channel (ID or URL) to a category.
  channel remove <category> <channel>
                                    Remove a YouTube channel from a category.
  feed <category>                   Retrieve the combined feed (last 20 videos) from a category.

TUI:
  submanager tui                    Launch the TUI frontend.

Example:
  submanager cli category create Music
  submanager cli channel add Music UC-9-kyTW8ZkZNDHQJ6FgpwQ
  submanager cli feed Music
  submanager tui
`
	fmt.Println(usage)
}
