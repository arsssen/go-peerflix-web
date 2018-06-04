package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

//torrentClient is the global Client instance
var torrentClient Client

type appConfig struct {
	HTTPPort       int
	TorrentPort    int
	Seed           bool
	TCP            bool
	MaxConnections int
	TempDir        string
}

//config is the global appConfig instance
var config appConfig

func main() {
	// Parse flags.
	flag.IntVar(&config.HTTPPort, "port", 8080, "http server port(ui and streaming)")
	flag.IntVar(&config.TorrentPort, "torrent-port", 5007, "Port to listen for incoming torrent connections")
	flag.BoolVar(&config.Seed, "seed", false, "Seed after finished downloading")
	flag.IntVar(&config.MaxConnections, "conn", 200, "Maximum number of connections")
	flag.BoolVar(&config.TCP, "tcp", true, "Allow connections via TCP")
	flag.StringVar(&config.TempDir, "dir", os.TempDir(), "Temporary directory path (default: /tmp)")
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir("ui")))
	http.HandleFunc("/download", handleDownload)
	http.HandleFunc("/stopdownload", handleStopDownload)
	http.HandleFunc("/play", handlePlay)
	http.HandleFunc("/playinomx", handlePlayInOmx)
	http.HandleFunc("/omxcmd", handleOmxCmd)
	http.HandleFunc("/status", handleStatus)
	http.HandleFunc("/stream", torrentClient.GetFile)
	http.HandleFunc("/search", handleSearch)
	fmt.Printf("starting web server on port %d\n", config.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.HTTPPort), nil))

}
