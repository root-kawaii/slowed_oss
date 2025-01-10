package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
	"io"
)

// RequestBody represents the structure of the incoming JSON payload
type RequestBody struct {
	YouTubeLink string `json:"youtube_link"`
}

func main() {
	http.HandleFunc("/submit", handleYouTubeLink)

	// Start the server on port 8080
	fmt.Println("Server is running on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// handleYouTubeLink handles the POST request containing a YouTube link
// handleYouTubeLink handles the POST request containing a YouTube link
func handleYouTubeLink(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    var requestBody RequestBody

    // Decode the JSON body of the request
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate that the YouTube link is not empty
    if requestBody.YouTubeLink == "" {
        http.Error(w, "YouTube link cannot be empty", http.StatusBadRequest)
        return
    }

    // Fetch the audio
    originalAudioPath, err := fetchYouTubeAudio(requestBody.YouTubeLink)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to fetch audio: %v", err), http.StatusInternalServerError)
        return
    }

    // Process the audio to make it slower
    slowedAudioPath, err := slowDownAudio(originalAudioPath)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to process audio: %v", err), http.StatusInternalServerError)
        return
    }

    // Add reverb to the slowed audio
    reverbAudioPath, err := addReverb(slowedAudioPath)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to add reverb: %v", err), http.StatusInternalServerError)
        return
    }

    // Stream the reverb audio file back to the client
    w.Header().Set("Content-Type", "audio/mpeg")
    w.Header().Set("Content-Disposition", "attachment; filename=reverb_audio.mp3")

    // Open the reverb audio file
    file, err := os.Open(reverbAudioPath)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to open reverb audio file: %v", err), http.StatusInternalServerError)
        return
    }
    defer file.Close()

    // Copy the file contents to the response writer
    if _, err := io.Copy(w, file); err != nil {
        http.Error(w, fmt.Sprintf("Failed to stream audio: %v", err), http.StatusInternalServerError)
        return
    }
}


// fetchYouTubeAudio downloads the audio from the provided YouTube link
func fetchYouTubeAudio(link string) (string, error) {
	// Generate a timestamp to include in the filename
	timestamp := time.Now().Format("20060102150405") // Format as: YYYYMMDDHHMMSS

	// Define the output file name with the timestamp
	outputFile := fmt.Sprintf("files/audio_%s.mp3", timestamp[8:14])

	// Construct the yt-dlp command
	cmd := exec.Command("yt-dlp", "--extract-audio", "--audio-format", "mp3", "-o", outputFile, link)

	// Set the command's output to os.Stdout for logging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("yt-dlp command failed: %v", err)
	}

	// Return the path to the downloaded audio
	return outputFile, nil
}

// slowDownAudio slows down the given audio file using FFmpeg
func slowDownAudio(inputPath string) (string, error) {
	// Generate a timestamp to include in the filename
	timestamp := time.Now().Format("20060102150405")

	// Define the output file name with the timestamp
	outputPath := fmt.Sprintf("files/audio_slowed_%s.mp3", timestamp[8:14])

	// Construct the FFmpeg command
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-filter:a", "asetrate=40000", "-vn", outputPath)

	// Set the command's output to os.Stdout for logging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg command failed: %v", err)
	}

	// Return the path to the slowed-down audio
	return outputPath, nil
}

// addReverb adds a reverb effect to the given audio file using FFmpeg
func addReverb(inputPath string) (string, error) {
	// Generate a timestamp to include in the filename
	timestamp := time.Now().Format("20060102150405")

	// Define the output file name with the timestamp
	outputPath := fmt.Sprintf("files/audio_reverb_%s.mp3", timestamp[8:14])

	// Construct the FFmpeg command
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-filter:a", "aecho=0.8:0.9:1000:0.3", "-vn", outputPath)

	// Set the command's output to os.Stdout for logging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg command failed: %v", err)
	}

	// Return the path to the audio with reverb
	return outputPath, nil
}
