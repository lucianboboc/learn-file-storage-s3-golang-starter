package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	base := make([]byte, 32)
	_, err := rand.Read(base)
	if err != nil {
		panic(err)
	}
	id := base64.RawURLEncoding.EncodeToString(base)
	return fmt.Sprintf("%s%s", id, getExtensnsion(mediaType))
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func getExtensnsion(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

type VideoAspectRatio struct {
	Streams []struct {
		DisplayAspectRatio string `json:"display_aspect_ratio"`
	} `json:"streams"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmdStr := fmt.Sprintf("ffprobe -v error -print_format json -show_streams %s", filePath)
	cmd := exec.Command("sh", "-c", cmdStr)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	if err := cmd.Run(); err != nil {
		return "", err
	}

	var res VideoAspectRatio
	if err := json.Unmarshal(buffer.Bytes(), &res); err != nil {
		return "", err
	}

	if len(res.Streams) == 0 {
		return "", errors.New("no streams found")
	}

	return res.Streams[0].DisplayAspectRatio, nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := filePath + ".processing"
	cmdStr := fmt.Sprintf("ffmpeg -i %s -c copy -movflags faststart -f mp4 %s", filePath, outputPath)
	cmd := exec.Command("sh", "-c", cmdStr)
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outputPath, nil
}
