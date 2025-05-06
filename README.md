# fshare
Restful service for uploading files and sharing them with others.

<br />
<br /> 
Behavior:<br />
- Each registered API key (user) receives its own folder<br />
- Duplicate files in a folder will be overwritten.<br />
<br />
<br /> 
Usage:<br />
When you start the service for the first time, make sure...:<br />
... the configured upload folder exists<br />
... neither the database file exists nor the home folder (name = uuid) is present in the upload folder
