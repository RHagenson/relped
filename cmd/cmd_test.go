package cmd_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestZeroExits(t *testing.T) {
	build := exec.Command("go", "install")
	if err := build.Run(); err != nil {
		t.Errorf("Failed to build relped: %s", err)
	}
	if relped, err := exec.LookPath("relped"); err == nil {
		minimumUse := exec.Command(relped,
			"build",
			"--relatedness=./example-data/relatedness-nums-and-codes.csv",
			"--output=/dev/null")
		wd, _ := os.Executable()
		minimumUse.Dir = wd
		fmt.Println(minimumUse.Dir)
		if out, err := minimumUse.CombinedOutput(); err != nil {
			t.Errorf("minimum use of relped failed: %s\n%s\n%s", relped, out, err)
		}
	} else {
		t.Errorf("Failed to find relped: %s", err)
	}

}
