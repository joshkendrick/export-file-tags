## Export File Tags

This is a small utility to export the tags I've put on my photos and videos to a sqlite database so they don't get lost (like when Windows changes something and I lose them, or it turns out there isn't a way to read them anymore, or...? I donno?)

Makes use of the [exiftool](https://exiftool.org) executable v12.65 (saved in this repo, and already outdated)

Also expects a sqlite database named media-tags.db in the same directory as the executable with the database structure already created. The sql to create the database is in media-tags.db.sql

Run the executable with a filepath
```
.\exporter_v0.1.0.exe <file-path-to-root-dir>
```

### Development

Due to the sqlite library, I had to install a gcc compiler, so with choco I installed mingw

Also had to set the CGO_ENABLED flag to 1:
```
// shows the full go environment
go env 

// displays CGO_ENABLED value
go env CGO_ENABLED

// set CGO_ENABLED
go env -w CGO_ENABLED=1
```

#### If you run this on your target machine, it should default to the correct $GOOS and $GOARCH to build:
```
go build -o exporter_v0.1.0.exe main.go
```
