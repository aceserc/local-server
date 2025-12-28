# Taranga AR Bug Hunt - APK Download Server

A lightweight, high-performance Go-based server designed for distributing APKs or assets during events. It features team verification, one-time download enforcement, and a sleek real-time admin dashboard.

## 🚀 Features

-   **Team Verification:** Access is restricted to teams pre-registered in the system via `teams.csv`.
-   **One-Time Download:** Prevents multiple downloads by the same team to ensure resource control and prevent unauthorized sharing.
-   **Dynamic File Distribution:** Supports serving both individual files (e.g., `.apk`) and entire directories.
-   **Real-time Admin Dashboard:** A premium, dark-themed dashboard providing:
    -   Live statistics (Total, Downloaded, and Pending teams).
    -   Detailed activity logs with timestamps and client IP addresses.
    -   Categorized views for completed and pending team syncs.
-   **Automatic Database Management:** Environment-ready with automatic SQLite database creation (`teams.db` and `downloads_log.db`).
-   **QR Code Utility:** Built-in tool to generate and display a QR code in the terminal for the server's network URL, facilitating quick access for mobile users.
-   **Network Awareness:** Automatically detects the host's local network IP to simplify mobile connectivity.
-   **Embedded UI:** All frontend assets are embedded into the Go binary for simple, single-file deployment.

## 🛠️ Tech Stack

-   **Backend:** Go (Optimized for 1.25+)
-   **Database:** SQLite3
-   **Frontend:** HTML5, CSS3 (Vanilla), Embedded via `go:embed`
-   **Design:** Premium Dark Mode (Google Fonts: Syne & Poppins)

## 📦 Getting Started

### 1. Configure Teams
Create or edit `teams.csv` in the root directory. The server uses this to seed the internal database.
```csv
Team Name, Team Number
Shadow Hunters, 101
Byte Masters, 102
```

### 2. Set Up the Server
Open `main.go` and adjust the constants to match your environment:
-   `adminPassphrase`: Security key for dashboard access.
-   `downloadPath`: Path to the file or folder you wish to distribute.
-   `port`: Default is `:8080`.

### 3. Run the Server
```bash
go mod tidy
go run main.go
```

### 4. Display the QR Code (Optional)
To help participants quickly access the download page, run the QR utility in a separate terminal:
```bash
go run utils/qr.go
```

## 🔐 Admin Dashboard

The dashboard is accessible via:
`http://localhost:8080/admin?pass=your_passphrase`

Replace `your_passphrase` with the value set in `main.go`.

## 📁 Project Structure

-   `main.go`: Main server application logic and API.
-   `utils/qr.go`: Utility to generate and display server URL as a QR code.
-   `index.html`: Public-facing verification and download page.
-   `admin.html`: Real-time administrative dashboard.
-   `teams.csv`: Source file for registered teams.
-   `teams.db`: Persistent storage for team data.
-   `downloads_log.db`: Persistent storage for download history and logs.
-   `download_server.log`: System log file for auditing.

## 🛡️ License

This project is open-source and available under the [MIT License](LICENSE).
