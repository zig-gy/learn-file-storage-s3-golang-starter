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
	// outputs: "16:9", "9:16" and "other"
	divided := float64(width) / float64(height)
	if divided == 16/9.0 {
		return "16:9"
	}
	if divided == 9/16.0 {
		return "9:16"
	}
	return "other"
}
