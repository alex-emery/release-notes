package input

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	title     string
	textInput textinput.Model
	body      *string
}

func Run(title string) (string, error) {
	body := new(string)
	p := tea.NewProgram(New(title, body))
	if _, err := p.Run(); err != nil {
		return "", err
	}

	return *body, nil
}

func New(title string, body *string) Model {
	ti := textinput.New()
	ti.Focus()
	ti.Width = 200

	return Model{
		title:     title,
		textInput: ti,
		body:      body,
	}

}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			value := m.textInput.Value()
			*m.body = value
			return m, tea.Quit
		}

		// We handle errors just like any other message
		// case errMsg:
		// 	m.err = msg
		// 	return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.title,
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
