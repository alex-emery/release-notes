package wizard

import (
	"fmt"

	"github.com/alex-emery/release-notes/internal/model/filter"
	"github.com/alex-emery/release-notes/internal/model/input"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"sigs.k8s.io/kustomize/api/types"
)

type Step int

const (
	EnvPage Step = iota
	NamespacePage
	ImagePage
	VersionPage
	DonePage
)

type Model struct {
	step         Step
	list         list.Model
	input        input.Model
	basepath     string
	environment  string
	namespace    string
	image        string
	imageVersion string
	kustFile     *types.Kustomization
	err          error
}

func NewModel(basepath string) Model {
	return Model{
		basepath: basepath,
		list:     filter.New("Select an environment", []string{"dev", "stage", "prod"}),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.step {
			case EnvPage:
				m.environment = string(m.list.SelectedItem().(filter.Item))
				namespaces, err := GetNamespaces(m.basepath, string(m.environment))
				if err != nil {
					m.err = fmt.Errorf("failed to get namespaces: %w", err)
				}

				m.list = filter.New("Select a namespace", namespaces)
				m.step = NamespacePage
				return m, nil

			case NamespacePage:
				m.namespace = string(m.list.SelectedItem().(filter.Item))
				kustFile, err := OpenKustomization(m.basepath, string(m.environment), string(m.namespace))
				if err != nil {
					m.err = fmt.Errorf("failed to open kustomization file: %w", err)
				}

				m.kustFile = kustFile

				images := make([]string, 0, len(kustFile.Images))
				for _, image := range kustFile.Images {
					images = append(images, image.Name)
				}

				m.list = filter.New("Select an image", images)
				m.step = ImagePage
				return m, nil
			case ImagePage:
				m.image = string(m.list.SelectedItem().(filter.Item))

				originalVersion := ""
				for _, image := range m.kustFile.Images {
					if image.Name == m.image {
						originalVersion = image.NewTag
						break
					}
				}

				m.input = input.New("Enter a new version", originalVersion)
				m.step = VersionPage
				return m, nil
			case VersionPage:
				m.imageVersion = m.input.Value()
				err := UpdateImageVersion(m.basepath, m.environment, m.namespace, m.image, m.imageVersion)
				if err != nil {
					m.err = fmt.Errorf("failed to update image version: %w", err)
				}

				m.step = DonePage
			}
		}
	}
	if m.step == VersionPage {
		newModel, cmd := m.input.Update(msg)
		m.input = newModel.(input.Model)
		return m, cmd
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %s", m.err.Error())
	}

	if m.step == VersionPage {
		return m.input.View()
	}
	return m.list.View()
}
