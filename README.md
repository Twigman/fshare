# fshare

A lightweight RESTful service to upload, share, view and delete files via UUID-based links.

## Features

- Simple REST API for file sharing
- API key–based user isolation
- Each API key gets a dedicated "home" folder
- Uploaded files are stored under `/upload/<user-uuid>/filename`
- File preview with syntax highlighting (for code/text files)
- Optional auto-deletion after a configurable number of hours (not yet implemented)

---

## Quick Start

### 1. Clone the project

```bash
git clone https://github.com/twigman/fshare.git
cd fshare
```

### 2. Prepare the upload directory

```bash
mkdir -p ./data/uploads
```

> Make sure the directory matches the path configured in your config file.

### 3. Generate a save API key

```bash
openssl rand -hex 32
```

> In the following steps the key `123` is used.

### 3. Run fshare

#### First start (creates initial API key):

```bash
go run ./src --config ./data/config.json --api-key 123 --comment "initial key" --highly-trusted
```

#### Subsequent starts:

```bash
go run ./src --config ./data/config.json
```

---

## Example Usage

### ✅ Upload a file

```bash
curl -X POST http://localhost:8080/upload \
     -H "Authorization: Bearer 123" \
     -F "file=@./data/config.json" \
     -F "is_private=false" \
     -F "auto_del_in_h=24"
```

**Response:**

```json
{"uuid": "0196af20-4ca0-7e02-9441-dfd94cd75b39"}
```

### 📎 Share link:

```
http://localhost:8080/r/0196af20-4ca0-7e02-9441-dfd94cd75b39
```

### 🗑 Delete a file:

```bash
curl -X DELETE http://localhost:8080/delete/0196af20-4ca0-7e02-9441-dfd94cd75b39 \
     -H "Authorization: Bearer 123"
```

---

## ⚙️ Configuration

The application can be configured using a JSON file. Below is a description of all available configuration fields.

## Configuration Fields

| Field                   | Type   | Description                                                                 |
|-------------------------|--------|-----------------------------------------------------------------------------|
| `port`                  | int    | Port the HTTP server will listen on                                         |
| `upload_path`           | string | Local directory where uploaded files are stored (must exist)                |
| `max_file_size_in_mb`   | int    | Maximum allowed size per file upload, in megabytes                          |
| `sqlite_db_path`        | string | Path to the SQLite database file used to store resources and API keys       |
| `autodelete_interval_in_sec`  | int    | Interval (in seconds) at which expired files (past their TTL) are automatically deleted            |

## 🏁 Command-Line Flags

In addition to configuration via file, the application also needs command-line flags for certain runtime parameters.

### Available Flags

| Flag              | Type    | Required | Description                                                                 |
|-------------------|---------|----------|-----------------------------------------------------------------------------|
| `--config`        | string  | ✅ yes   | Path to the JSON configuration file                                         |
| `--api-key`       | string  | ⛔ optional* | Initial API key to bootstrap the system (first start)                   |
| `--comment`       | string  | ⛔ optional | Optional comment describing the initial API key                          |
| `--highly-trusted`| bool    | ⛔ optional | Grants elevated privileges to the initial API key user                   |

### Notes

- `--config` must always be provided; the application will not start without it.
- If `--api-key` is provided, the system will attempt to create a new key on startup.
- `--comment` and `--highly-trusted` are only relevant when `--api-key` is used.
- Files uploaded by users with the `--highly-trusted` flag may be rendered directly in the browser, even if the file type could potentially contain active or unsafe content (pdf, svg, etc.). Additionally, trusted API keys are allowed to create new API keys via the dedicated endpoint.
