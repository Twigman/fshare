# fshare
Restful service for uploading files and sharing them with others.
<br /> 
Behavior:<br />
- Each registered API key (user) receives its own folder.<br />
- Duplicate files will be overwritten.<br />
- No folders allowed at this point.<br />
<br /> 
Preparation:<br />
Before you start the service for the first time, make sure...:<br />
... the configured upload folder exists<br />
... neither the database file exists nor the home folder (name = uuid) is present in the upload folder
<br />

## Usage
Navigate in /fshare/src.<br />
First start (pass config + initial API key with comment):

```
go run . --config "../data/config.json" --api-key 123 --comment 123
```

Following starts:

```
go run . --config "../data/config.json"
```
