# fshare

A lightweight RESTful service to upload, share, view and delete files via UUID-based links.

## Features

- Simple REST API for file sharing
- API key–based user isolation
- Each API key gets a dedicated "home" folder
- Uploaded files are stored under `/<upload-folder>/<apikey-uuid>/<filename>`
- File preview with syntax highlighting (for code/text files)
- Configurable time to live (TTL) for every uploaded file

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
     -F "auto_del_in=2h"
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


## 📥 Upload Endpoint Parameters

### Request Headers

| Header          | Type    | Required | Description                          |
|-----------------|---------|----------|--------------------------------------|
| `Authorization` | string  | ✅       | Bearer token (`Bearer <API-Key>`)    |

### Form Data (`multipart/form-data`)

| Field          | Type    | Required | Description                                                                                                                                           | Example         |
|----------------|---------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------|
| `file`         | file    | ✅       | The file to upload.                                                                                                                                   | `myfile.txt`    |
| `is_private`   | boolean | ❌       | Whether the file is private. Accepts `true` or `false`. Defaults to `false`.                                                                          | `true`          |
| `auto_del_in`  | string  | ❌       | Time to live (TTL) for the file. Can be a duration (e.g., `24h`, `30m`) or days (e.g., `2d`). If omitted, the file does not expire automatically.     | `2d`, `24h`, `30m` |

---

## 🔑 API-Key Management Endpoint

### POST /api-key

Create a new API key. Requires a valid **highly trusted** API key for authorization.

#### Request Headers

| Header          | Type    | Required | Description                          |
|-----------------|---------|----------|--------------------------------------|
| `Authorization` | string  | ✅       | Bearer token (`Bearer <API-Key>`)    |
| `Content-Type`  | string  | ✅       | Must be `application/json`           |

#### Request JSON Body

| Field            | Type    | Required | Description                                          | Example             |
|------------------|---------|----------|------------------------------------------------------|---------------------|
| `key`            | string  | ✅       | The API key value to create                          | `my-new-api-key`    |
| `comment`        | string  | ❌       | Optional comment for the API key                      | `test key`          |
| `highly_trusted` | boolean | ❌       | Whether the key should have elevated privileges       | `false`             |

#### Response (201 Created)

```json
{
  "uuid": "9b8a71c2-1234-4567-8910-abcdef123456",
  "comment": "optional description",
  "highly_trusted": false,
  "created_at": "2024-05-27T12:34:56Z"
}
```

#### Example Request (cURL)

```bash
curl -X POST http://localhost:8080/api-key \
     -H "Authorization: Bearer <trusted-api-key>" \
     -H "Content-Type: application/json" \
     -d '{
           "key": "my-new-api-key",
           "comment": "optional description",
           "highly_trusted": false
         }'
```

---

## ⚙️ Configuration

The application can be configured using a JSON file. Below is a description of all available configuration fields.

## Configuration Fields (config.json)

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
- Files uploaded by users with the `--highly-trusted` flag may be rendered directly in the browser, even if the file type could potentially contain active or unsafe content (pdf, svg). Additionally, trusted API keys are allowed to create new API keys via the dedicated endpoint.
