package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed index.html
var indexHTML string

//go:embed admin.html
var adminHTML string

var publicTmpl = template.Must(template.New("index").Parse(indexHTML))
var adminTmpl = template.Must(template.New("admin").Parse(adminHTML))

var teamsDB *sql.DB
var logDB *sql.DB
var fileLogger *log.Logger

const (
	csvFile         = "teams.csv"
	downloadPath    = "download.txt" // ← CHANGE TO YOUR ACTUAL APK/FILE
	logFileName     = "download_server.log"
	port            = ":8080"
	adminPassphrase = "sandip" // ← CHANGE THIS TO YOUR SECRET PASSWORD
)

func initLogger() {
	file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Cannot create log file:", err)
	}
	fileLogger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fileLogger.Printf("Warning: Could not determine local IP: %v", err)
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func main() {
	initLogger()
	fileLogger.Println("=== Taranga AR Bug Hunt Server Started ===")
	defer fileLogger.Println("=== Server Stopped ===")

	var err error
	teamsDB, err = sql.Open("sqlite3", "teams.db")
	if err != nil {
		fileLogger.Fatal("Failed to open teams.db:", err)
	}
	defer teamsDB.Close()

	logDB, err = sql.Open("sqlite3", "downloads_log.db")
	if err != nil {
		fileLogger.Fatal("Failed to open downloads_log.db:", err)
	}
	defer logDB.Close()

	if err = teamsDB.Ping(); err != nil {
		fileLogger.Fatal("teamsDB ping failed:", err)
	}
	if err = logDB.Ping(); err != nil {
		fileLogger.Fatal("logDB ping failed:", err)
	}

	createTables()
	seedFromCSV()

	localIP := getLocalIP()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/verify", verifyAndDownloadHandler)
	http.HandleFunc("/admin", adminAuth(adminHandler))

	fmt.Println("\n🌟 Taranga AR Bug Hunt - APK Download Server")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("Public Download Page : http://localhost:8080")
	if localIP != "127.0.0.1" {
		fmt.Printf("Network Access       : http://%s:8080\n", localIP)
	}
	fmt.Println("\n🔐 Admin Dashboard   : http://localhost:8080/admin?pass=" + adminPassphrase)
	fmt.Printf("   Passphrase: %s\n", adminPassphrase)
	fmt.Println("\nLog file:", logFileName)
	fmt.Println("Waiting for teams... 🐞")

	fileLogger.Printf("Server running on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// Middleware to protect /admin
func adminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getIP(r)
		pass := r.URL.Query().Get("pass")

		if pass != adminPassphrase {
			fileLogger.Printf("Unauthorized admin access attempt from %s", clientIP)
			http.Redirect(w, r, "/?error=admin_unauthorized", http.StatusSeeOther)
			return
		}

		fileLogger.Printf("Admin panel accessed by %s", clientIP)
		next(w, r)
	}
}

func createTables() {
	// ... (same as before)
	_, err := teamsDB.Exec(`
		CREATE TABLE IF NOT EXISTS teams (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			team_name TEXT NOT NULL,
			team_number INTEGER NOT NULL,
			UNIQUE(team_name, team_number)
		);
	`)
	if err != nil {
		fileLogger.Fatal("Failed to create teams table:", err)
	}

	_, err = logDB.Exec(`
		CREATE TABLE IF NOT EXISTS download_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			team_name TEXT NOT NULL,
			team_number INTEGER NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			ip_address TEXT,
			downloaded BOOLEAN DEFAULT 0,
			UNIQUE(team_name, team_number)
		);
	`)
	if err != nil {
		fileLogger.Fatal("Failed to create download_logs table:", err)
	}

	fileLogger.Println("Database tables ready")
}

func seedFromCSV() {
	file, err := os.Open(csvFile)
	if err != nil {
		fileLogger.Printf("Warning: %s not found – no teams seeded (%v)", csvFile, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fileLogger.Printf("Error reading CSV: %v", err)
		return
	}
	if len(records) == 0 {
		fileLogger.Println("CSV is empty")
		return
	}

	start := 0
	if len(records) > 0 && strings.Contains(strings.ToLower(records[0][0]), "team") {
		start = 1
	}

	tx, err := teamsDB.Begin()
	if err != nil {
		fileLogger.Printf("Transaction begin failed: %v", err)
		return
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO teams (team_name, team_number) VALUES (?, ?)")
	if err != nil {
		fileLogger.Printf("Prepare statement failed: %v", err)
		tx.Rollback()
		return
	}
	defer stmt.Close()

	seeded := 0
	for i := start; i < len(records); i++ {
		row := records[i]
		if len(row) < 2 {
			continue
		}
		name := strings.TrimSpace(row[0])
		numStr := strings.TrimSpace(row[1])
		num, err := strconv.Atoi(numStr)
		if err != nil || name == "" || num <= 0 {
			continue
		}

		if _, err := stmt.Exec(name, num); err == nil {
			fileLogger.Printf("Seeded team: %s (#%d)", name, num)
			seeded++
		}
	}

	if err := tx.Commit(); err != nil {
		fileLogger.Printf("Commit failed: %v", err)
	} else {
		fileLogger.Printf("Successfully seeded %d teams from CSV", seeded)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Show admin unauthorized message if someone tries to access /admin without pass
	if strings.HasPrefix(r.URL.Path, "/admin") {
		// Let the adminAuth middleware handle it (it will redirect)
		http.Redirect(w, r, "/?error=admin_unauthorized", http.StatusSeeOther)
		return
	}

	// Serve the public download page
	w.Header().Set("Content-Type", "text/html")
	if err := publicTmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		fileLogger.Println("Public template error:", err)
	}
}
func verifyAndDownloadHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := getIP(r)

	if r.Method != http.MethodPost {
		fileLogger.Printf("Invalid request (not POST) from %s", clientIP)
		http.Redirect(w, r, "/?error=invalid", http.StatusSeeOther)
		return
	}

	teamName := strings.TrimSpace(r.FormValue("team_name"))
	teamNumStr := strings.TrimSpace(r.FormValue("team_number"))

	teamNumber, err := strconv.Atoi(teamNumStr)
	if err != nil || teamName == "" || teamNumber <= 0 {
		fileLogger.Printf("Invalid input from %s: name='%s', number='%s'", clientIP, teamName, teamNumStr)
		http.Redirect(w, r, "/?error=invalid", http.StatusSeeOther)
		return
	}

	fileLogger.Printf("Verification attempt: %s (#%d) from %s", teamName, teamNumber, clientIP)

	// Check if team is registered
	var count int
	err = teamsDB.QueryRow("SELECT COUNT(*) FROM teams WHERE team_name = ? AND team_number = ?", teamName, teamNumber).Scan(&count)
	if err != nil || count == 0 {
		fileLogger.Printf("Unauthorized access attempt: %s (#%d) from %s – team not found", teamName, teamNumber, clientIP)
		http.Redirect(w, r, "/?error=unauthorized", http.StatusSeeOther)
		return
	}

	// Check if already downloaded
	var alreadyDownloaded bool
	err = logDB.QueryRow("SELECT downloaded FROM download_logs WHERE team_name = ? AND team_number = ?", teamName, teamNumber).Scan(&alreadyDownloaded)
	if err == nil && alreadyDownloaded {
		fileLogger.Printf("Blocked re-download: %s (#%d) from %s – already downloaded", teamName, teamNumber, clientIP)
		http.Redirect(w, r, "/?error=already_downloaded", http.StatusSeeOther)
		return
	}

	// Allow download
	if err != nil { // No previous log
		_, err = logDB.Exec("INSERT INTO download_logs (team_name, team_number, ip_address, downloaded) VALUES (?, ?, ?, 1)",
			teamName, teamNumber, clientIP)
	} else {
		_, err = logDB.Exec("UPDATE download_logs SET downloaded = 1, timestamp = CURRENT_TIMESTAMP, ip_address = ? WHERE team_name = ? AND team_number = ?",
			clientIP, teamName, teamNumber)
	}

	if err != nil {
		fileLogger.Printf("Failed to log download for %s (#%d): %v", teamName, teamNumber, err)
	}

	fileLogger.Printf("✅ DOWNLOAD GRANTED: %s (#%d) from %s at %s", teamName, teamNumber, clientIP, time.Now().Format("2006-01-02 15:04:05"))

	// Serve file or folder
	fileInfo, err := os.Stat(downloadPath)
	if err != nil {
		fileLogger.Printf("Resource not found: %s", downloadPath)
		http.Redirect(w, r, "/?error=invalid", http.StatusSeeOther)
		return
	}

	if fileInfo.IsDir() {
		fileLogger.Printf("Serving folder: %s to %s (#%d)", downloadPath, teamName, teamNumber)
		http.FileServer(http.Dir(downloadPath)).ServeHTTP(w, r)
	} else {
		fileLogger.Printf("Serving file: %s (as attachment) to %s (#%d)", downloadPath, teamName, teamNumber)
		w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(downloadPath))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, downloadPath)
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	data := loadDashboardData()

	w.Header().Set("Content-Type", "text/html")
	if err := adminTmpl.Execute(w, data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		fileLogger.Println("Admin template error:", err)
	}
}

type Team struct {
	Name   string
	Number int
}

type DownloadLog struct {
	TeamName   string
	TeamNumber int
	Timestamp  string
	IP         string
}

type DashboardData struct {
	TotalTeams          int
	DownloadedTeams     int
	PendingTeams        int
	RegisteredTeams     []Team
	DownloadedTeamsList []Team
	PendingTeamsList    []Team
	DownloadLogs        []DownloadLog
	GeneratedAt         string
}

func loadDashboardData() DashboardData {
	data := DashboardData{
		GeneratedAt: time.Now().Format("January 2, 2006 15:04 MST"),
	}

	// Registered teams
	rows, _ := teamsDB.Query("SELECT team_name, team_number FROM teams ORDER BY team_name")
	defer rows.Close()
	for rows.Next() {
		var t Team
		rows.Scan(&t.Name, &t.Number)
		data.RegisteredTeams = append(data.RegisteredTeams, t)
	}
	data.TotalTeams = len(data.RegisteredTeams)

	// Downloaded teams
	rows, _ = logDB.Query("SELECT team_name, team_number FROM download_logs WHERE downloaded = 1 ORDER BY timestamp DESC")
	defer rows.Close()
	for rows.Next() {
		var t Team
		rows.Scan(&t.Name, &t.Number)
		data.DownloadedTeamsList = append(data.DownloadedTeamsList, t)
	}
	data.DownloadedTeams = len(data.DownloadedTeamsList)

	// Pending
	downloadedMap := make(map[string]bool)
	for _, dt := range data.DownloadedTeamsList {
		downloadedMap[fmt.Sprintf("%s-%d", dt.Name, dt.Number)] = true
	}
	for _, rt := range data.RegisteredTeams {
		if !downloadedMap[fmt.Sprintf("%s-%d", rt.Name, rt.Number)] {
			data.PendingTeamsList = append(data.PendingTeamsList, rt)
		}
	}
	data.PendingTeams = len(data.PendingTeamsList)

	// Logs
	rows, _ = logDB.Query("SELECT team_name, team_number, timestamp, ip_address FROM download_logs WHERE downloaded = 1 ORDER BY timestamp DESC")
	defer rows.Close()
	for rows.Next() {
		var l DownloadLog
		var ts time.Time
		rows.Scan(&l.TeamName, &l.TeamNumber, &ts, &l.IP)
		l.Timestamp = ts.Format("Jan 2, 15:04")
		data.DownloadLogs = append(data.DownloadLogs, l)
	}

	return data
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	return ip
}
