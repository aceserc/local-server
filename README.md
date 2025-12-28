# Taranga AR Bug Hunt - APK Download Server

A lightweight, secure Go-based server designed for distributing APKs (or other files) during events. It features team verification, one-time download enforcement, and a real-time admin dashboard.

## 🚀 Features

-   **Team Verification:** Only teams registered in the system (via `teams.csv`) can access the download.
-   **One-Time Download:** Prevents multiple downloads by the same team to ensure fair play or resource control.
-   **Admin Dashboard:** Real-time statistics including:
    -   Total registered teams.
    -   Successfully downloaded teams.
    -   Pending teams.
    -   Detailed download logs with IP addresses and timestamps.
-   **SQLite Backend:** Simple and portable database for storing team data and download logs.
-   **Embedded UI:** HTML templates are embedded into the Go binary for easy deployment.
-   **Detailed Logging:** All activities are logged to both the console and a file (`download_server.log`).

## 🛠️ Prerequisites

-   **Go:** Version 1.23 or higher.
-   **GCC:** (Required by `go-sqlite3` for CGO)

## 📦 Getting Started

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/dev-sandip/local-server.git
    cd local-server
    ```

2.  **Prepare your team list:**
    Edit `teams.csv` to include your participating teams. The format should be:
    ```csv
    Team Name, Team Number
    Team Alpha, 101
    Team Beta, 102
    ```

3.  **Configure the server (Optional):**
    Open `main.go` and modify the constants at the top:
    -   `adminPassphrase`: The password for accessing the admin panel.
    -   `port`: The port the server will run on (default `:8080`).
    -   `downloadPath`: Path to the APK or folder you want to distribute.

4.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

5.  **Run the server:**
    ```bash
    go run main.go
    ```

## 🔐 Admin Dashboard

Access the admin dashboard at:
`http://localhost:8080/admin?pass=your_passphrase`

Replace `your_passphrase` with the value set in `main.go` (default is `sandip`).

## 📁 Project Structure

-   `main.go`: The core server logic and API handlers.
-   `index.html`: The public-facing team verification and download page.
-   `admin.html`: The administrative dashboard.
-   `teams.csv`: Source file for seeding the registered teams list.
-   `teams.db`: SQLite database storing registered team information.
-   `downloads_log.db`: SQLite database tracking download history.
-   `download_server.log`: Text file for persistent server logs.

## 🛡️ License

This project is open-source and available under the [MIT License](LICENSE).
