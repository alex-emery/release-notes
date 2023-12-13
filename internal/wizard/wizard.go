package wizard

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/api/types"
)

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
