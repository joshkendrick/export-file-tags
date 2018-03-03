## Export File Tags

This is a small utility to export the tags I've placed on my photos and videos and back them up to a boltDB database so they don't get lost (like when Windows changes something and I lose them, or it turns out there isn't a way to read them anymore... I don't know... it makes me feel better?)

Makes use of the [exiftool](https://www.sno.phy.queensu.ca/~phil/exiftool/) executable v10.81 (saved in this repo, and already outdated)

I built the Windows exporter executable from reading [this](https://github.com/golang/go/wiki/WindowsCrossCompiling)

#### If you run this on your target machine, it will default to the correct $GOOS and $GOARCH:
```
go build -o exporter_v0.0.2.exe main.go
```
