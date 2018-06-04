# go-peerflix-web

Start watching the movie while your torrent is still downloading.
Control torrent and playback remotely(in browser) via web interface.
Integrated search to find movies on [rutor](http://www.rutor.is)

When installed on Raspberry PI, can play torrents in omxplayer.

Also provides a stream that can be played in any player like VLC or directly in browser.

## Building / installing

just run  `go build`

Building for raspberry:

`GOARCH=arm go build`

Then copy go-peerflix-web and the `ui` directory to your Raspberry and run it there.

## Usage

just run `go-peerflix-web` and open corresponding url(e.g. http://localhost:8080) in your browser.

### Command-line flags

- port  - specify web server port. Default is 8080.
- dir - directory to store downloaded and temporary files. Default: os temp dir(/tmp for linux)
- torrent-port - specify port for incoming torrent connections. Default is 5007
- seed - to continue seeding after download is completed. Default: false.
- conn - maximum number of connections. Default 200

## Used third-party sources

- [Jquery](https://jquery.com/)
- [Pure CSS](https://purecss.io/)
- Some code snippets from [go-peerflix](https://github.com/Sioro-Neoku/go-peerflix)