package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command(
		"swag", "init",
		"-g", "cmd/swagger/main.go",
		"-o", "docs/api",
		"--parseDependency",
		"--parseInternal",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Generating swagger docs...")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
