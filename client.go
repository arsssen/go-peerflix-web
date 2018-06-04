package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/iplist"
	"github.com/dustin/go-humanize"
)

const torrentBlockListURL = "http://john.bitsurge.net/public/biglist.p2p.gz"

var isHTTP = regexp.MustCompile(`^https?:\/\/`)

// Client manages the torrent downloading.
type Client struct {
	Started       bool
	Downloading   bool
	TorrentClient *torrent.Client
	Torrent       *torrent.Torrent
	Progress      int64
	Uploaded      int64
}

func (c *Client) startTorrent(url string) (err error) {
	// Create client.
	c.TorrentClient, err = torrent.NewClient(&torrent.Config{
		DataDir:    config.TempDir,
		NoUpload:   !config.Seed,
		Seed:       config.Seed,
		DisableTCP: !config.TCP,
		ListenAddr: fmt.Sprintf(":%d", config.TorrentPort),
	})
	c.Started = true
	if err != nil {
		fmt.Printf("err adding torrent %#v", err)
		return err
	}

	// Add torrent.
	log.Printf("adding torrent %#v", url)
	// Add as magnet url.
	if strings.HasPrefix(url, "magnet:") {
		if c.Torrent, err = c.TorrentClient.AddMagnet(url); err != nil {
			return err
		}
	} else {
		// Otherwise add as a torrent file.

		// If it's online, we try downloading the file.
		if isHTTP.MatchString(url) {
			if url, err = downloadFile(url); err != nil {
				return err
			}
		}

		if c.Torrent, err = c.TorrentClient.AddTorrentFromFile(url); err != nil {
			return err
		}
	}

	c.Torrent.SetMaxEstablishedConns(config.MaxConnections)

	go func() {
		<-c.Torrent.GotInfo()

		c.Downloading = true
		log.Printf("got info %#v", c)
		c.Torrent.DownloadAll()

		// Prioritize first 2% of the file.
		c.getLargestFile().DownloadRegion(0, int64(c.Torrent.NumPieces()/100*2))
	}()

	go c.addBlocklist()

	return
}

// Download and add the blocklist.
func (c *Client) addBlocklist() {
	var err error
	blocklistPath := config.TempDir + "/go-peerflix-web-blocklist.gz"

	if _, err = os.Stat(blocklistPath); os.IsNotExist(err) {
		err = downloadBlockList(blocklistPath)
	}

	if err != nil {
		log.Printf("Error downloading blocklist: %s", err)
		return
	}

	// Load blocklist.
	blocklistReader, err := os.Open(blocklistPath)
	if err != nil {
		log.Printf("Error opening blocklist: %s", err)
		return
	}

	// Extract file.
	gzipReader, err := gzip.NewReader(blocklistReader)
	if err != nil {
		log.Printf("Error extracting blocklist: %s", err)
		return
	}

	// Read as iplist.
	blocklist, err := iplist.NewFromReader(gzipReader)
	if err != nil {
		log.Printf("Error reading blocklist: %s", err)
		return
	}

	log.Printf("Loading blocklist.\nFound %d ranges\n", blocklist.NumRanges())
	c.TorrentClient.SetIPBlockList(blocklist)
}

func downloadBlockList(blocklistPath string) (err error) {
	log.Printf("Downloading blocklist")
	fileName, err := downloadFile(torrentBlockListURL)
	if err != nil {
		log.Printf("Error downloading blocklist: %s\n", err)
		return
	}

	return os.Rename(fileName, blocklistPath)
}

// Close cleans up the connections.
func (c *Client) Close() {
	if c.Torrent != nil {
		c.Torrent.Drop()
	}
	if c.TorrentClient != nil {
		c.TorrentClient.Close()
	}
	c.Started = false
	c.Downloading = false
	c.Progress = 0
	c.Uploaded = 0
}

//Status is a struct containing torrent status
type Status struct {
	Name          string `json:"name"`
	Stream        string `json:"stream"`
	Progress      string `json:"progress"`
	DownloadSpeed string `json:"down"`
	UploadSpeed   string `json:"up"`
	Started       bool   `json:"started"`
	Downloading   bool   `json:"downloading"`
	OmxPlaying    bool   `json:"omx_playing"`
}

// Status returns the status
func (c *Client) Status() (st Status) {

	st.Downloading = c.Downloading
	st.Started = c.Started

	if !c.Downloading || c.Torrent.Info() == nil {
		return
	}

	currentProgress := c.Torrent.BytesCompleted()
	downloadSpeed := humanize.Bytes(uint64(currentProgress-c.Progress)) + "/s"
	c.Progress = currentProgress

	complete := humanize.Bytes(uint64(currentProgress))
	size := humanize.Bytes(uint64(c.Torrent.Info().TotalLength()))

	uploadProgress := c.Torrent.Stats().DataBytesWritten - c.Uploaded
	uploadSpeed := humanize.Bytes(uint64(uploadProgress)) + "/s"
	c.Uploaded = uploadProgress

	st.Name = c.Torrent.Info().Name

	if c.ReadyForPlayback() {
		st.Stream = fmt.Sprintf("/:%d/stream", config.HTTPPort)
	}
	if currentProgress > 0 {
		st.Progress = fmt.Sprintf("%s/%s %.2f%%", complete, size, c.percentage())
	}
	if currentProgress < c.Torrent.Info().TotalLength() {
		st.DownloadSpeed = downloadSpeed
	}
	if config.Seed {
		st.UploadSpeed = uploadSpeed
	}
	return
}

func (c Client) getLargestFile() *torrent.File {
	var target torrent.File
	var maxSize int64
	for _, file := range c.Torrent.Files() {
		if maxSize < file.Length() {
			maxSize = file.Length()
			target = file
		}
	}
	return &target
}

// ReadyForPlayback checks if the torrent is ready for playback or not.
// We wait until 2% of the torrent to start playing.
func (c Client) ReadyForPlayback() bool {
	return c.percentage() > 2
}

// GetFile is an http handler to serve the biggest file managed by the client.
func (c *Client) GetFile(w http.ResponseWriter, r *http.Request) {

	target := c.getLargestFile()
	entry, err := NewFileReader(target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("\nGetFile %#v %v\n", target, entry)
	defer func() {
		if err := entry.Close(); err != nil {
			log.Printf("Error closing file reader: %s\n", err)
		}
	}()

	w.Header().Set("Content-Disposition", "attachment; filename=\""+c.Torrent.Info().Name+"\"")
	http.ServeContent(w, r, target.DisplayPath(), time.Now(), entry)
}

func (c Client) percentage() float64 {
	info := c.Torrent.Info()

	if info == nil {
		return 0
	}

	return float64(c.Torrent.BytesCompleted()) / float64(info.TotalLength()) * 100
}

func downloadFile(URL string) (fileName string, err error) {
	var file *os.File
	if file, err = ioutil.TempFile(config.TempDir, "go-peerflix-web"); err != nil {
		return
	}

	defer func() {
		if ferr := file.Close(); ferr != nil {
			log.Printf("Error closing torrent file: %s", ferr)
		}
	}()

	response, err := http.Get(URL)
	if err != nil {
		return
	}

	defer func() {
		if ferr := response.Body.Close(); ferr != nil {
			log.Printf("Error closing torrent file: %s", ferr)
		}
	}()

	_, err = io.Copy(file, response.Body)

	return file.Name(), err
}

// SeekableContent describes an io.ReadSeeker that can be closed as well.
type SeekableContent interface {
	io.ReadSeeker
	io.Closer
}

// FileEntry helps reading a torrent file.
type FileEntry struct {
	*torrent.File
	torrent.Reader
}

// Seek seeks to the correct file position, paying attention to the offset.
func (f FileEntry) Seek(offset int64, whence int) (int64, error) {
	return f.Reader.Seek(offset+f.File.Offset(), whence)
}

// NewFileReader sets up a torrent file for streaming reading.
func NewFileReader(f *torrent.File) (SeekableContent, error) {
	torrent := f.Torrent()
	reader := torrent.NewReader()

	// We read ahead 1% of the file continuously.
	reader.SetReadahead(f.Length() / 100)
	reader.SetResponsive()
	_, err := reader.Seek(f.Offset(), os.SEEK_SET)

	return &FileEntry{
		File:   f,
		Reader: reader,
	}, err
}
