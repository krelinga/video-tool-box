# video-tool-box
A collection of tools useful for ripping, transcoding, or otherwise processing media.

## Running TCServer on Windows

```
set SMB_USER="<smb username>"
set SMB_PASSWORD="<smb password>"
curl.exe "https://raw.githubusercontent.com/krelinga/video-tool-box/main/tcserver-compose.yaml" | docker compose --file - up -d --pull always
```
