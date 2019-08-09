package osutil

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	log.Printf("Running:\n%s %s\n", cmd.Path, strings.Join(cmd.Args, " "))

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running external cmd %s (%s)", cmd.Path, err)
	}

	return nil
}
