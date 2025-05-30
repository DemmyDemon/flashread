package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const DEBUG = false

type TickMsg struct {
	time time.Time
	id   int
}

type model struct {
	width    int
	Filename string
	File     *os.File
	Scanner  *bufio.Scanner
	Speed    int
	Word     int
	Line     []string
	Message  string
	Paused   bool
	ticker   int
}

func getModel(filename string) model {
	var scanner *bufio.Scanner
	var file *os.File = nil
	var err error

	if filename != "" {
		file, err = os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open file: %v", err)
			os.Exit(1)
		}
		scanner = bufio.NewScanner(file)
	} else {
		filename = "STDIN"
		file = nil
		scanner = bufio.NewScanner(os.Stdin)
	}

	return model{
		Filename: filename,
		Scanner:  scanner,
		File:     file,
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
		m.Message = "There is no Scanner"
		m.Paused = true
		return
	}

	if !m.Scanner.Scan() {

		if m.File == nil {
			m.Message = "No file: Releasing Scanner"
			m.Scanner = nil
		} else {
			_, err := m.File.Seek(0, 0)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to seek to start of file: %v\n", err)
				os.Exit(1)
			}
			m.Message = "Resetting Scanner"
			m.Scanner = bufio.NewScanner(m.File)
		}

		m.Line = []string{"End of file."}
		m.Paused = true
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
func (m model) tickCmd() tea.Msg {
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
			m.Speed++
		case "shift+up":
			m.Speed += 10
		case "down":
			m.Speed--
			if m.Speed <= 0 {
				m.Speed = 1
			}
		case "shift+down":
			m.Speed -= 10
			if m.Speed <= 0 {
				m.Speed = 1
			}
		default:
			m.Message = msg.String()
		}
	case TickMsg:
		if msg.id != m.ticker {
			return m, nil
		}
		if !m.Paused {
			m.nextWord()
		}
		speed := time.Duration(60_000_000_000 / m.Speed) // Words per minute, in nanoseconds
		m.ticker++
		return m, tea.Tick(speed, m.tick)
	}
	return m, nil
}

func (m model) View() string {
	speed := fmt.Sprintf("Paused (Space to resume at %dwpm)", m.Speed)
	if !m.Paused {
		speed = fmt.Sprintf("Speed: %dwpm (Space to pause)", m.Speed)
	}
	word := m.Line[m.Word]
	padwidth := (m.width / 2) - (len(word) / 2)
	if padwidth <= 0 {
		padwidth = 0
	}
	word = fmt.Sprintf("%s%s", strings.Repeat(" ", padwidth), word)
	if DEBUG {
		return fmt.Sprintf("Reading %s\n\n%s\n\n%s\n%s", m.Filename, word, speed, m.Message)
	}
	return fmt.Sprintf("Reading %s\n\n%s\n\n%s", m.Filename, word, speed)
}

func main() {
	filename := ""
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}
	p := tea.NewProgram(getModel(filename))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Holy moly: %v", err)
		os.Exit(1)
	}
}
