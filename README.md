# fshare

A lightweight RESTful service to upload, share, view and delete files via UUID-based links.

## Features

- Simple REST API for file sharing
- API key–based user isolation
- Each API key gets a dedicated "home" folder
- Uploaded files are stored under `/upload/<user-uuid>/filename`
- File preview with syntax highlighting (for code/text files)
- Optional auto-deletion after a configurable number of hours
- Soft delete support (`deleted_at` timestamp)

---

## Behavior

- Each API key maps to one user-specific home directory.
- Only file uploads are currently supported – folders are not yet handled.

---

## Quick Start

### 1. Clone the project

```bash
git clone https://github.com/YOUR_USERNAME/fshare.git
cd fshare
```

### 2. Prepare the upload directory

```bash
mkdir -p ./data/uploads
```

> Make sure the directory matches the path configured in your config file.

### 3. Run fshare

#### First start (creates initial API key):

```bash
go run ./src --config ./data/config.json --api-key 123 --comment "initial key"
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

## Configuration

> Coming soon (optional: list of available config fields).

---

## License

MIT © Twigman
