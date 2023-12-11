package input_test

import (
	"fmt"
	"testing"

	"github.com/alex-emery/release-notes/pkg/input"
)

func TestInput(t *testing.T) {
	answer, _ := input.Run("test")
	fmt.Println(answer)
}
