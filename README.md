# fshare
Restful service to share files.<br />

## Description

This tool provides simple REST endpoints to share files via links.  A separate "home" directory is created in the configured upload folder for each registered API key. These Uploads are automaticlly placed in the corresponding <br />

Available Parameters:



Behavior:<br />
- Each registered API key (user) receives its own folder.<br />
- Duplicate files will be overwritten.<br />
- No folders allowed at this point.<br />

## Preparation

Before you start the service for the first time, make sure...<br />
... the configured upload folder exists<br />
... neither the database file exists nor the home folder (name = uuid) is present in the upload folder

## Getting Started
1. Checkout the project and navigate in /fshare<br />
2. Create /data/uploads:

```
mkdir ./data/uploads
```

3. Start fshare:

```
# first start with initial API key
go run ./src --config "./data/config.json" --api-key 123 --comment 123

# following starts
go run ./src --config "./data/config.json"
```

4. Upload a file:

```
curl -X POST http://localhost:8080/upload \
     -H "Authorization: Bearer 123" \
     -F "file=@./data/config.json" \
     -F "is_private=false" \
     -F "auto_del_in_h=24"

# response
# {"uuid":"0196af20-4ca0-7e02-9441-dfd94cd75b39"}
```

5. Visit the upload: localhost:8080/v/0196af20-4ca0-7e02-9441-dfd94cd75b39<br />

6. Delete a file:

```
curl -X POST http://localhost:8080/delete/0196af20-4ca0-7e02-9441-dfd94cd75b39 \
     -H "Authorization: Bearer 123"
```
