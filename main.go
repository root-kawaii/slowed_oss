package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// RequestBody represents the structure of the incoming JSON payload
type RequestBody struct {
	YouTubeLink string `json:"youtube_link"`
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the server!")
	})

	http.HandleFunc("/submit", handleYouTubeLink)

	port := os.Getenv("PORT") // Get PORT from Render
	if port == "" {
		port = "10000" // Default to 8080 locally
	}

	// Start the server on port 8080
	fmt.Println("Server is running on port " + port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

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

	// Process the audio and stream the output while processing
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Disposition", "attachment; filename=reverb_audio.mp3")

	// Start processing audio (slow it down and add reverb) while streaming
	cmd := exec.Command("ffmpeg", "-i", originalAudioPath, "-filter:a", "asetrate=40000, aecho=0.8:0.9:1000:0.1", "-vn", "-f", "mp3", "pipe:1")
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	// Start the processing
	if err := cmd.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start ffmpeg: %v", err), http.StatusInternalServerError)
		return
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		http.Error(w, fmt.Sprintf("Error while processing audio: %v", err), http.StatusInternalServerError)
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
