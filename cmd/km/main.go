package main

import (
	"github.com/bsycorp/keymaster/cmd/km/commands"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

func main() {
	commands.Execute()
}

func WriteFile(data []byte, localPath string, perm os.FileMode) {
	log.Printf("Writing local file: %s", localPath)
	err := ioutil.WriteFile(localPath, data, perm)
	if err != nil {
		log.Fatalf("Failed to write local file: %s: %s", localPath, err)
	}
	if err := FixWindowsPerms(localPath); err != nil {
		log.Fatalf("Failed to set file permissions: %s: %s", localPath, err)
	}
}
