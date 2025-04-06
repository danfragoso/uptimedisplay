package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Check if a file path is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path-to-png-file>")
		os.Exit(1)
	}

	// Get the file path from command line argument
	filePath := os.Args[1]

	// Check if file has .png extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".png") {
		fmt.Println("Error: File must be a PNG image")
		os.Exit(1)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Decode PNG
	img, err := png.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding PNG: %v\n", err)
		os.Exit(1)
	}

	// Get image bounds
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// Create RGB565BE byte array (2 bytes per pixel)
	bytes := make([]byte, width*height*2)

	// Convert image to RGB565BE format
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			// Convert from uint32 (0-65535) to appropriate bit ranges:
			// R: 5 bits (0-31), G: 6 bits (0-63), B: 5 bits (0-31)
			r5 := byte((r >> 11) & 0x1F) // 5 bits for red
			g6 := byte((g >> 10) & 0x3F) // 6 bits for green
			b5 := byte((b >> 11) & 0x1F) // 5 bits for blue

			// Calculate index in the bytes array
			// Each pixel takes 2 bytes
			index := ((y-bounds.Min.Y)*width + (x - bounds.Min.X)) * 2

			// Pack into RGB565 format (Big Endian)
			// First byte: RRRRRGGG
			bytes[index] = (r5 << 3) | (g6 >> 3)
			// Second byte: GGGBBBBB
			bytes[index+1] = ((g6 & 0x07) << 5) | b5
		}
	}

	// Generate output filename based on input filename
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	outputFileName := "../" + baseName + "_rgb565be.go"

	// Create output file
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	// Write package declaration and imports
	outputFile.WriteString("package main\n\n")

	// Write dimensions variables
	fmt.Fprintf(outputFile, "var imgWidth = %d\n", width)
	fmt.Fprintf(outputFile, "var imgHeight = %d\n\n", height)

	// Write byte array with proper Go syntax
	outputFile.WriteString("var imgBytes = []byte{\n\t")

	// Write bytes with formatting
	for i, b := range bytes {
		fmt.Fprintf(outputFile, "0x%02x, ", b)

		// Add line breaks for readability (16 bytes per line)
		if (i+1)%16 == 0 && i < len(bytes)-1 {
			outputFile.WriteString("\n\t")
		}
	}

	outputFile.WriteString("\n}\n")

	fmt.Printf("Successfully converted '%s' to RGB565BE format\n", filePath)
	fmt.Printf("Image dimensions: %d x %d pixels\n", width, height)
	fmt.Printf("Output saved to '%s'\n", outputFileName)
}
