# fshare
Restful service to share files.<br />
<br /> 
Behavior:<br />
- Each registered API key (user) receives its own folder.<br />
- Duplicate files will be overwritten.<br />
- No folders allowed at this point.<br />

## Preparation

Before you start the service for the first time, make sure...:<br />
... the configured upload folder exists<br />
... neither the database file exists nor the home folder (name = uuid) is present in the upload folder

## Getting Started
1. Checkout the project and navigate in /fshare<br />
2. Create /data/uploads:

```
mkdir ./data/uploads
```

3. For the first start pass config + initial API key (and comment):

```
go run ./src --config "./data/config.json" --api-key 123 --comment 123
```

3.1. Following starts:

```
go run ./src --config "./data/config.json"
```

4. Upload a file:

```
curl -X POST http://localhost:8080/upload \
     -H "Authorization: Bearer 123" \
     -F "file=@./data/config.json" \
     -F "is_private=false" \
     -F "auto_del_in_h=24"
```
