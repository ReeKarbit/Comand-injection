package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {
	// Path folder tempat file Java berada
	directory := "./"
	logPath := "fix_log.txt"

	// Buka file log untuk mencatat perubahan
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Gagal membuka file log:", err)
		return
	}
	defer logFile.Close()

	fmt.Println("🔍 Memeriksa semua file Java di:", directory)

	// Loop semua file dalam direktori
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".java") {
			processFile(path, logFile)
		}
		return nil
	})

	if err != nil {
		log.Println("Terjadi kesalahan saat membaca folder:", err)
	}
}

// Fungsi untuk mendeteksi dan memperbaiki command injection dalam satu file
func processFile(filePath string, logFile *os.File) {
	tempPath := filePath + ".tmp"
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Gagal membuka file:", filePath)
		return
	}
	defer file.Close()

	tempFile, err := os.Create(tempPath)
	if err != nil {
		log.Println("Gagal membuat file sementara:", tempPath)
		return
	}
	defer tempFile.Close()

	// Regex untuk mendeteksi pola command injection
	reRuntime := regexp.MustCompile(`Runtime\.getRuntime\(\)\.exec\((.*?)\)`)
	reProcessBuilder := regexp.MustCompile(`new ProcessBuilder\("sh", "-c", (.*?)\)`)

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	changes := 0
	logEntries := []string{}

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		updatedLine := line
		modified := false

		// Deteksi `Runtime.getRuntime().exec()`
		if reRuntime.MatchString(line) {
			fmt.Printf("\n🔴 Kerentanan ditemukan di %s (Baris %d):\n%s\n", filePath, lineNumber, line)

			// Saran perbaikan
			fixedLine := strings.Replace(line, "Runtime.getRuntime().exec(", "new ProcessBuilder(", 1)

			fmt.Println("\n💡 Saran perbaikan:")
			fmt.Println(fixedLine)
			fmt.Print("\n✔ Terima perbaikan ini? (y/n): ")

			var response string
			fmt.Scanln(&response)

			if strings.ToLower(response) == "y" {
				updatedLine = fixedLine
				modified = true
			}
		}

		// Deteksi `ProcessBuilder("sh", "-c", "... " + userInput)`
		if reProcessBuilder.MatchString(line) {
			fmt.Printf("\n🔴 Kerentanan ditemukan di %s (Baris %d):\n%s\n", filePath, lineNumber, line)

			// Saran perbaikan: gunakan argumen array + validasi input
			fixedLine := `if (!userInput.matches("^[a-zA-Z0-9._-]+$")) {
    System.out.println("Input tidak valid!");
    return;
}
ProcessBuilder pb = new ProcessBuilder("ls", userInput);`

			fmt.Println("\n💡 Saran perbaikan:")
			fmt.Println(fixedLine)
			fmt.Print("\n✔ Terima perbaikan ini? (y/n): ")

			var response string
			fmt.Scanln(&response)

			if strings.ToLower(response) == "y" {
				updatedLine = fixedLine
				modified = true
			}
		}

		// Simpan hasil perbaikan atau kode asli
		tempFile.WriteString(updatedLine + "\n")

		// Catat perubahan ke log
		if modified {
			logEntries = append(logEntries, fmt.Sprintf("%s (Baris %d):\nSebelum: %s\nSesudah: %s\n",
				filePath, lineNumber, line, updatedLine))
			changes++
		}
	}

	// Jika ada perubahan, perbarui file asli
	if changes > 0 {
		os.Rename(tempPath, filePath)
		fmt.Printf("\n✅ Perubahan diterapkan pada %s\n", filePath)

		// Simpan log perubahan dengan timestamp
		logFile.WriteString(fmt.Sprintf("\n[%s] Perbaikan diterapkan pada file %s\n",
			time.Now().Format("2006-01-02 15:04:05"), filePath))
		for _, entry := range logEntries {
			logFile.WriteString(entry + "\n")
		}
	} else {
		os.Remove(tempPath)
		fmt.Printf("\nℹ Tidak ada perubahan di %s\n", filePath)
	}
}
