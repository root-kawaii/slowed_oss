package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// RequestBody represents the structure of the incoming JSON payload
type RequestBody struct {
	YouTubeLink string `json:"youtube_link"`
}

func main() {
	// Get the current directory of the Go program
	execDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// Paths to the certificate and key files in the current directory
	certFile := filepath.Join(execDir, "cert.pem")
	keyFile := filepath.Join(execDir, "key.pem")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the HTTPS server!")
	})

	http.HandleFunc("/submit", handleYouTubeLink)

	port := os.Getenv("PORT") // Get PORT from Render/Railway/etc.
	if port == "" {
		port = "8080" // Default for local testing
	}

	// HTTPS Server Configuration
	server := &http.Server{
		Addr:    ":" + port,
		Handler: nil, // Default handlers
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12, // Enforce modern TLS
		},
	}

	fmt.Println("🚀 HTTPS Server is running on port " + port)

	// Check if the certificate and key files exist in the current directory
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		fmt.Println("⚠️ No SSL certs found! Generate self-signed certs with:")
		fmt.Println("   openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes")
		os.Exit(1)
	}

	// Start the HTTPS server
	err = server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		log.Fatalf("Error starting HTTPS server: %v", err)
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
