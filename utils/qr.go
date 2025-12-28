package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

const (
	port      = "8080"
	eventDate = "January 12, 2026"
)

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("Warning: Could not detect local IP: %v (using localhost)", err)
		return "localhost"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func printQRLeftAligned(qr *qrcode.QRCode) {
	bitmap := qr.Bitmap()
	matrixSize := len(bitmap)
	// size := matrixSize + 8 // Quiet zone: 4 modules each side

	fmt.Println("\n") // Small top margin

	for y := -4; y < matrixSize+4; y++ {
		var line strings.Builder
		for x := -4; x < matrixSize+4; x++ {
			if x >= 0 && x < matrixSize && y >= 0 && y < matrixSize {
				if bitmap[y][x] {
					line.WriteString("██")
				} else {
					line.WriteString("  ")
				}
			} else {
				line.WriteString("  ") // Quiet zone
			}
		}
		fmt.Println(line.String())
	}

	fmt.Println("") // Bottom margin
}

func main() {
	ip := getLocalIP()
	url := fmt.Sprintf("http://%s:%s", ip, port)

	if ip == "localhost" {
		fmt.Println("Warning: Network IP not detected — using localhost only")
		url = fmt.Sprintf("http://localhost:%s", port)
	}

	fmt.Printf("🌟 Taranga AR Bug Hunt — %s\n", eventDate)
	fmt.Println("════════════════════════════════════")
	fmt.Println("📱 Participants: Scan the QR code below to download the APK")

	// Generate QR code
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		log.Fatal("Failed to generate QR code:", err)
	}

	// Display left-aligned
	printQRLeftAligned(qr)

	fmt.Printf("🔗 URL: %s\n\n", url)
	fmt.Println("Teams can now scan this QR code from their phones!")
	fmt.Println("Good luck finding those bugs! 🐞")
}
