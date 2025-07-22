package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func processVideoForFastStart(filePath string) (string, error) {
	newPath := filePath + ".processing"
	fastStartCommand := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", newPath)
	outBytes := bytes.NewBuffer([]byte{})
	fastStartCommand.Stdout = outBytes
	if err := fastStartCommand.Run(); err != nil {
		return "", fmt.Errorf("could not decode output: %v", err)
	}

	return newPath, nil
}
