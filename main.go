package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type TickMsg struct {
	time time.Time
	id   int
}

type model struct {
	width    int
	Filename string
	Source   io.ReadCloser
	Scanner  *bufio.Scanner
	Speed    int
	Word     int
	Line     []string
	Message  string
	Paused   bool
	ticker   int
}

func getModel(filename string) model {
	file, err := os.Open(filename)
	scanner := bufio.NewScanner(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file: %v", err)
		os.Exit(1)
	}
	return model{
		Filename: filename,
		Scanner:  scanner,
		Source:   file,
		Speed:    250,
		Word:     0,
		Message:  "",
		Line:     []string{"Press space to begin."},
		Paused:   true,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) nextLine() {
	if m.Scanner == nil {
		return
	}

	if !m.Scanner.Scan() {
		err := m.Source.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Aw shucks: %v", err)
			os.Exit(1)
		}
		m.Scanner = nil
		m.Line = []string{"End", "of", "file"}
		return
	}

	rawLine := m.Scanner.Text()
	m.Line = strings.Split(rawLine, " ")
	if len(m.Line) == 0 {
		m.Line = []string{""}
	}
}

func (m *model) nextWord() {
	m.Word++
	if m.Word >= len(m.Line) {
		m.Word = 0
		m.nextLine()
	}
}

func (m model) tick(_ time.Time) tea.Msg {
	return m.tickCmd()
}
func (m *model) tickCmd() tea.Msg {
	m.ticker++
	return TickMsg{
		time: time.Now(),
		id:   m.ticker,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.Message = fmt.Sprintf("width: %d", msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case " ":
			m.Paused = !m.Paused
			if !m.Paused {
				return m, m.tickCmd
			}
		case "tab":
			m.nextWord()
		case "up":
			if !m.Paused {
				m.Speed++
			}
		case "down":
			m.Speed--
			if m.Speed <= 0 {
				m.Speed = 1
				m.Paused = true
			}
		default:
			m.Message = msg.String()
		}
	case TickMsg:
		if msg.id != m.ticker && m.ticker != 0 {
			m.Message = fmt.Sprintf("%d != %d", msg.id, m.ticker)
			return m, nil
		}
		if !m.Paused {
			m.nextWord()
		}
		speed := time.Duration(60_000_000_000 / m.Speed) // Words per minute, in nanoseconds
		return m, tea.Tick(speed, m.tick)
	}
	return m, nil
}

func (m model) View() string {
	speed := "Paused (Space to resume)"
	if !m.Paused {
		speed = fmt.Sprintf("Speed: %d (Space to pause)", m.Speed)
	}
	word := m.Line[m.Word]
	padwidth := (m.width / 2) - (len(word) / 2)
	if padwidth <= 0 {
		padwidth = 0
	}
	word = fmt.Sprintf("%s%s", strings.Repeat(" ", padwidth), word)
	return fmt.Sprintf("Reading %s\n\n%s\n\n%s\n%s", m.Filename, word, speed, m.Message)
}

func main() {
	p := tea.NewProgram(getModel("LICENSE"))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Holy moly: %v", err)
		os.Exit(1)
	}
}
