package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// RequestBody represents the structure of the incoming JSON payload
type RequestBody struct {
	YouTubeLink string `json:"youtube_link"`
}

func main() {
	// Get PORT from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to 8080 if not set
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the server!")
	})

	http.HandleFunc("/submit", handleYouTubeLink)

	// Start the server on the correct port
	fmt.Println("Server is running on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
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

	// Process the audio (slow down and add reverb)
	processedAudio, err := processAudio(originalAudioPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process audio: %v", err), http.StatusInternalServerError)
		return
	}

	// Set response headers and stream the processed audio
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Disposition", "attachment; filename=processed_audio.wav")
	w.Write(processedAudio)
}

// fetchYouTubeAudio downloads the audio from the provided YouTube link
func fetchYouTubeAudio(link string) (string, error) {
	// Generate a timestamp to include in the filename
	timestamp := time.Now().Format("20060102150405") // Format as: YYYYMMDDHHMMSS

	// Ensure the files directory exists
	err := os.MkdirAll("files", os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create files directory: %v", err)
	}

	// Define the output file name with the timestamp
	outputFile := fmt.Sprintf("files/audio_%s.wav", timestamp[8:14])

	// Create a simple dummy WAV file with a sine wave for testing purposes
	err = createDummyWavFile(outputFile)
	if err != nil {
		return "", fmt.Errorf("failed to create dummy WAV file: %v", err)
	}

	// Returning the path to the "downloaded" audio
	return outputFile, nil
}

// createDummyWavFile generates a simple WAV file with a sine wave for testing
func createDummyWavFile(filePath string) error {
	// Define the parameters for the sine wave
	const sampleRate = 44100
	const duration = 2      // 2 seconds
	const frequency = 440.0 // Frequency of the sine wave (A4)

	// Create the output file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create a WAV encoder
	encoder := wav.NewEncoder(file, sampleRate, 16, 1, 1)

	// Generate the sine wave samples
	numSamples := sampleRate * duration
	samples := make([]int, numSamples)

	for i := 0; i < numSamples; i++ {
		// Generate the sine wave sample (simple formula)
		sample := int(32767 * 0.5 * (1 + math.Sin(2*math.Pi*frequency*float64(i)/float64(sampleRate))))
		samples[i] = sample
	}

	// Write the samples to the WAV file
	buffer := &audio.IntBuffer{Data: samples, Format: &audio.Format{SampleRate: sampleRate, NumChannels: 1}}
	err = encoder.Write(buffer)
	if err != nil {
		return fmt.Errorf("failed to write samples to WAV file: %v", err)
	}

	// Close the encoder and the file
	err = encoder.Close()
	if err != nil {
		return fmt.Errorf("failed to close WAV encoder: %v", err)
	}

	return nil
}

// processAudio applies a slowdown and reverb effect to the given audio file
func processAudio(audioPath string) ([]byte, error) {
	// Open the audio file (assuming WAV format)
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %v", err)
	}
	defer file.Close()

	// Decode the WAV audio file
	decoder := wav.NewDecoder(file)
	audioData, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to decode WAV file: %v", err)
	}

	// Slow down the audio by skipping every other sample (simple method)
	slowdownFactor := 2 // Slow down by 2x (simple method)
	slowedDownData := &audio.IntBuffer{
		Format: audioData.Format,
		Data:   make([]int, len(audioData.Data)/slowdownFactor),
	}

	// Skip every other sample to simulate slowdown (this is a simple approach)
	for i := 0; i < len(slowedDownData.Data); i++ {
		slowedDownData.Data[i] = audioData.Data[i*slowdownFactor]
	}

	// Apply a simple reverb (echo) effect - basic example
	// This is a simplified approach to simulate reverb by adding previous samples
	for i := 1; i < len(slowedDownData.Data); i++ {
		// Echo: Current sample + a scaled previous sample
		slowedDownData.Data[i] += slowedDownData.Data[i-1] / 2
	}

	// Convert processed data to a new WAV file
	outputFile := fmt.Sprintf("processed_%s.wav", time.Now().Format("20060102150405"))
	outputFileHandler, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %v", err)
	}
	defer outputFileHandler.Close()

	// Write processed audio back to file
	encoder := wav.NewEncoder(outputFileHandler, 44100, 16, 1, 1)
	err = encoder.Write(slowedDownData)
	if err != nil {
		return nil, fmt.Errorf("failed to write processed audio: %v", err)
	}

	// Read the processed file into memory to stream back to the client
	processedAudio, err := ioutil.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read processed audio: %v", err)
	}

	// Return the processed audio file
	return processedAudio, nil
}
