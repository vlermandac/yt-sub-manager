package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	"github.com/yourusername/submanager/internal/manager"
)

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "quit"),
)

type errMsg error

type viewState int

const (
	viewCategories viewState = iota
	viewFeed
)

type model struct {
	spinner     spinner.Model
	state       viewState
	categories  []string
	selectedCat int
	feed        []manager.Video
	feedOffset  int
	quitting    bool
	err         error
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	cats, err := manager.GetCategories()
	if err != nil {
		cats = []string{}
	}

	return model{
		spinner:     s,
		state:       viewCategories,
		categories:  cats,
		selectedCat: 0,
		feed:        nil,
		feedOffset:  0,
	}
}

func (m model) Init() tea.Cmd {
	// Start the spinner tick.
	return tea.Batch(m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}

		switch m.state {
		case viewCategories:
			switch msg.String() {
			case "up", "k":
				if m.selectedCat > 0 {
					m.selectedCat--
				}
			case "down", "j":
				if m.selectedCat < len(m.categories)-1 {
					m.selectedCat++
				}
			case "enter":
				if len(m.categories) > 0 {
					cat := m.categories[m.selectedCat]
					feed, err := manager.GetFeedForCategory(cat)
					if err != nil {
						m.err = err
					} else {
						m.feed = feed
						m.state = viewFeed
						m.feedOffset = 0
					}
				}
			}
		case viewFeed:
			switch msg.String() {
			case "up", "k":
				if m.feedOffset > 0 {
					m.feedOffset--
				}
			case "down", "j":
				if m.feedOffset < len(m.feed)-1 {
					m.feedOffset++
				}
			case "b":
				m.state = viewCategories
			}
		}
		return m, nil

	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	var s string
	switch m.state {
	case viewCategories:
		s = "Categories:\n\n"
		if len(m.categories) == 0 {
			s += "No categories available. Use the CLI to add categories.\n"
		}
		for i, cat := range m.categories {
			cursor := "  "
			if i == m.selectedCat {
				cursor = "➜ "
			}
			s += fmt.Sprintf("%s%s\n", cursor, cat)
		}
		s += "\nPress Enter to view feed, or " + quitKeys.Help().Desc + ".\n"
	case viewFeed:
		s = "Feed:\n\n"
		if len(m.feed) == 0 {
			s += "No videos found.\n"
		} else {
			end := m.feedOffset + 5
			if end > len(m.feed) {
				end = len(m.feed)
			}
			for i := m.feedOffset; i < end; i++ {
				video := m.feed[i]
				pubDate := video.UploadDate
				if t, err := time.Parse("20060102", video.UploadDate); err == nil {
					pubDate = t.Format("2006-01-02")
				}
				s += fmt.Sprintf("Title: %s\nPublished: %s\nURL: %s\n\n", video.Title, pubDate, video.WebpageURL)
			}
		}
		s += "\nPress b to go back, or " + quitKeys.Help().Desc + ".\n"
	}
	// Prepend spinner view (as in the template)
	return "\n" + m.spinner.View() + s
}

// RunTUI starts the Bubble Tea program.
func RunTUI() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running TUI:", err)
		// Exit with error code.
		os.Exit(1)
	}
}
