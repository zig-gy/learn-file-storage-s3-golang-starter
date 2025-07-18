package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	type ffprobeIn struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	cmdOutput := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	outBytes := bytes.NewBuffer([]byte{})
	cmdOutput.Stdout = outBytes
	if err := cmdOutput.Run(); err != nil {
		fmt.Println("read")
		return "", fmt.Errorf("error running command: %v", err)
	}

	probeData := ffprobeIn{}
	decoder := json.NewDecoder(outBytes)
	if err := decoder.Decode(&probeData); err != nil {
		fmt.Println("decode")
		return "", fmt.Errorf("could not decode the output: %v", err)
	}

	return aspectRatioFromHeightWidth(probeData.Streams[0].Height, probeData.Streams[0].Width), nil
}

func aspectRatioFromHeightWidth(height, width int) string {
	// outputs: "16:9", "9:16" and "other"
	// fmt.Println(height, width)

	tolerance := 10
	if abs(width*9-height*16) <= tolerance {
		return "16:9"
	}
	if abs(width*16-height*9) <= tolerance {
		return "9:16"
	}
	return "other"
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
