## Export File Tags

This is a small utility to export the tags I've put on my photos and videos to a sqlite database so they don't get lost (like when Windows changes something and I lose them, or it turns out there isn't a way to read them anymore..., or...?)

Makes use of the [exiftool](https://exiftool.org) executable v12.65 (saved in this repo, and already outdated)

Also expects a sqlite database named media-tags.db on the same path with the database structure already created. The sql to create the database is in media-tags.db.sql

I built the Windows exporter executable from reading [this](https://github.com/golang/go/wiki/WindowsCrossCompiling)

Due to sqlite, I had to install a gcc compiler and set the CGO_ENABLED flag to 1
```
// shows the full go environment
go env 

// displays CGO_ENABLED value
go env CGO_ENABLED
```

#### If you run this on your target machine, it will default to the correct $GOOS and $GOARCH:
```
go build -o exporter_v0.0.2.exe main.go
```
