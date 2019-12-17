package version

import (
	"fmt"
	"os/exec"
	"strings"
)

var (
	GitTag string
)

func init() {
	cmd := exec.Command("git", "describe", "--tags")
	if tag, err := cmd.CombinedOutput(); err == nil {
		GitTag = strings.TrimSpace(string(tag))
	} else {
		panic(fmt.Sprintf("Could not set GitTag: %s\n", err))
	}
}
