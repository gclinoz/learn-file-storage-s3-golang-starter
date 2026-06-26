package main

import (
	"os/exec"
	"bytes"
	"encoding/json"
)

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command(
		"ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath,
	)

	output := bytes.Buffer{}  
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	type parameters struct {
		Streams []struct {
			Width	int `json:"width"`
			Height	int `json:"height"`
		} `json:"streams"`
	}

	params := parameters{}
	err = json.Unmarshal(output.Bytes(), &params)
	if err != nil {
		return "", err
	}
	
	ratio := params.Streams[0].Width / params.Streams[0].Height
	switch ratio {
	case 1:
		return "16:9", nil
	case 0:
		return "9:16", nil
	default:
		return "other", nil
	}
}
