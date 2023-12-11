package wizard

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/api/types"
)

func New(basepath string) *tview.Application {
	app := tview.NewApplication()

	errPage := tview.NewTextView()

	// Step 1: Pick an environment.
	envPage := tview.NewList().
		AddItem("dev", "", '1', nil).
		AddItem("stage", "", '2', nil).
		AddItem("prod", "", '3', nil)

	// Step 2: Pick a namespace.
	nsPage := tview.NewList()

	// Step 3: Pick an image to update.
	imagePage := tview.NewList()

	// Step 4: Input a new version to update to.
	imageVersionPage := tview.NewInputField()

	// Step 5: Repeat?
	confirmation := tview.NewModal().
		SetText("Update more images?").
		AddButtons([]string{"Y", "N"})

	// Set up the pages
	pages := tview.NewPages().
		AddPage("env", envPage, true, true).
		AddPage("namespace", nsPage, true, false).
		AddPage("image", imagePage, true, false).
		AddPage("version", imageVersionPage, true, false).
		AddPage("confirm", confirmation, true, false).
		AddPage("error", errPage, true, false)

	var selectedEnv = new(string)
	envPage.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		clean := strings.TrimSpace(mainText)
		selectedEnv = &clean

		namespaces, err := GetNamespaces(basepath, *selectedEnv)
		if err != nil {
			errPage.SetText(err.Error())
			pages.SwitchToPage("error")
			return
		}

		for _, ns := range namespaces {
			nsPage.AddItem(ns, "", 0, nil)
		}

		pages.SwitchToPage("namespace")
	})

	var selectedNs = new(string)
	var selectedKustFile = new(types.Kustomization)
	nsPage.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		clean := strings.TrimSpace(mainText)
		selectedNs = &clean

		kustFile, err := OpenKustomization(basepath, *selectedEnv, *selectedNs)
		if err != nil {
			errPage.SetText(err.Error())
			pages.SwitchToPage("error")
			return
		}

		selectedKustFile = kustFile
		for _, image := range kustFile.Images {
			imagePage.AddItem(image.Name, "", 0, nil)
		}

		pages.SwitchToPage("image")
	})

	var selectedImage = new(string)
	imagePage.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		clean := strings.TrimSpace(mainText)
		selectedImage = &clean
		imageVersionPage.SetTitle(*selectedImage)

		for _, image := range selectedKustFile.Images {
			if image.Name == *selectedImage {
				imageVersionPage.SetText(image.NewTag)
				break
			}
		}
		pages.SwitchToPage("version")
	})

	imageVersionPage.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			if err := UpdateImageVersion(basepath, *selectedEnv, *selectedNs, *selectedImage, imageVersionPage.GetText()); err != nil {
				errPage.SetText(err.Error())
				pages.SwitchToPage("error")
				return
			}

			pages.SwitchToPage("confirm")
		}
	})

	confirmation.
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Y" {
				pages.SwitchToPage("env")
			} else {

				app.Stop()
			}
		})

	app.SetRoot(pages, true)

	return app
}

func readMapSliceFromFile(filepath string) (yaml.MapSlice, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
	}

	m := yaml.MapSlice{}
	err = yaml.Unmarshal(file, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", filepath, err)
	}

	return m, nil
}

func writeMapSliceToFile(filepath string, m yaml.MapSlice) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal file %s: %w", filepath, err)
	}

	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filepath, err)
	}

	return nil
}

// finds the version field for a given image in a kustomization.yaml file
// and returns a pointer to it.
func findImageVersionInMapSlice(m *yaml.MapSlice, image, version string) *interface{} {
	for i, item := range *m {
		if item.Key == "images" {
			for j, images := range item.Value.([]interface{}) {
				images := images.(yaml.MapSlice)
				found := false
				for _, field := range images {
					if field.Key.(string) == "name" {
						if field.Value.(string) == image {
							found = true
							break
						}
					}
				}
				if found {
					for k, field := range images {
						if field.Key.(string) == "newTag" {

							return &(*m)[i].Value.([]interface{})[j].(yaml.MapSlice)[k].Value

						}
					}
				}
			}
		}
	}

	return nil
}

// updateImageInMapSlice updates the image version in a kustomization.yaml file.
// MapSlice maintains order, so we reduce noise in the diff when we write the file back out.
// This means we have to manually step through everything.
func updateImageInMapSlice(m *yaml.MapSlice, image, version string) bool {
	field := findImageVersionInMapSlice(m, image, version)
	if field == nil {
		return false
	}

	*field = version

	return true

}

func updateImageVersion(filepath, image, version string) error {
	m, err := readMapSliceFromFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	didUpdate := updateImageInMapSlice(&m, image, version)
	if !didUpdate {
		return fmt.Errorf("failed to update image %s to version %s", image, version)
	}

	err = writeMapSliceToFile(filepath, m)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filepath, err)
	}

	return nil
}

func UpdateImageVersion(base, env, ns, image, version string) error {
	filepath := path.Join(base, fmt.Sprintf("environments/engine-%s/baseline/%s/kustomization.yaml", env, ns))

	return updateImageVersion(filepath, image, version)
}

func OpenKustomization(base, env, ns string) (*types.Kustomization, error) {
	filepath := path.Join(base, fmt.Sprintf("environments/engine-%s/baseline/%s/kustomization.yaml", env, ns))

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath, err)
	}

	k := &types.Kustomization{}
	err = yaml.Unmarshal(data, k)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", filepath, err)
	}

	return k, nil
}

// GetNamspaces returns all folders that match the pattern:
// environments/engine-<env>/baseline/wb-*
func GetNamespaces(basepath, env string) ([]string, error) {
	directories, err := filepath.Glob(path.Join(basepath, fmt.Sprintf("environments/engine-%s/baseline/wb-*", env)))
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace directories: %w", err)
	}

	var namespaces []string
	for _, dir := range directories {
		fields := strings.Split(dir, "/")
		namespaces = append(namespaces, fields[len(fields)-1])
	}

	return namespaces, nil
}
