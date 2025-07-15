package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	type ffprobeIn struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	cmdOutput := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "show_streams", filePath)
	outBytes := bytes.NewBuffer([]byte{})
	cmdOutput.Stdout = outBytes
	if err := cmdOutput.Run(); err != nil {
		return "", fmt.Errorf("error running command: %v", err)
	}

	probeData := ffprobeIn{}
	decoder := json.NewDecoder(outBytes)
	if err := decoder.Decode(&probeData); err != nil {
		return "", fmt.Errorf("could not decode the output: %v", err)
	}

	return "", nil
}

func aspectRatioFromHeightWidth(height, width int) string {
	return ""
}
