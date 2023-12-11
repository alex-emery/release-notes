package wizard

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadWrite(t *testing.T) {
	m, err := readMapSliceFromFile("./kustomization.yaml")
	require.NoError(t, err)

	didUpdate := updateImageInMapSlice(&m, "some-image", "9.9.9")
	require.True(t, didUpdate)

	field := findImageVersionInMapSlice(&m, "some-image", "9.9.9")
	require.Equal(t, "9.9.9", *field)
}
