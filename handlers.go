package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

func handleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := r.ParseForm(); err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error": "parse request %s"}`, err)))
		return
	}
	results := searchRutor(r.Form.Get("search"))

	ret, err := json.Marshal(results)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error": "search cannot marshal %s"}`, err)))
		return
	}
	w.Write(ret)

}

func handleStream(w http.ResponseWriter, r *http.Request) {
	if torrentClient.Downloading {
		torrentClient.GetFile(w, r)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !torrentClient.Started {
		w.Write([]byte("{}"))
		return
	}
	status := torrentClient.Status()
	status.OmxPlaying = omxPlayer.Playing
	if resp, err := json.Marshal(status); err == nil {
		w.Write(resp)
	} else {
		w.Write([]byte(`{"error": "cant encode status into JSON"}`))
	}

}

func handlePlay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseForm(); err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error": "parse request %s"}`, err)))
		return
	}

	player := r.Form.Get("player")
	url := r.Form.Get("url")

	if player != "" && torrentClient.Downloading {
		go func() {
			for !torrentClient.ReadyForPlayback() {
				time.Sleep(time.Second)
			}
			command := []string{}
			if runtime.GOOS == "darwin" {
				command = []string{"open", "-a"}
			}
			command = append(command, player)
			command = append(command, url)
			exec.Command(command[0], command[1:]...).Start()
		}()
		w.Write([]byte(fmt.Sprintf(`{"ok": "playing %s in %s"}`, url, player)))
		return
	}
	w.Write([]byte(`{"error": "cant play"}`))
}

func handlePlayInOmx(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	url := fmt.Sprintf("http://localhost:%d/stream", config.HTTPPort)
	log.Printf("playing %s in omx", url)
	err := omxPlayer.Start(url)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
		return
	}
	w.Write([]byte(fmt.Sprintf(`{"ok": "playing %s"}`, url)))
}

func handleOmxCmd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseForm(); err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error": "parse request %s"}`, err)))
		return
	}
	cmd := r.Form.Get("cmd")
	if !omxPlayer.Playing {
		w.Write([]byte(`{"error": "omx is not started"}`))
		return
	}
	if err := omxPlayer.SendCommand(cmd); err != nil {
		w.Write([]byte(fmt.Sprintf(`{"cmd error": "%s"}`, err)))
		return
	}
	w.Write([]byte(fmt.Sprintf(`{"ok": " cmd %s sent to omx"}`, cmd)))
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseForm(); err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error": "parse request %s"}`, err)))
		return
	}
	fmt.Printf("download got %#v", r.Form.Get("url"))

	torrentClient = Client{}
	err := torrentClient.startTorrent(r.Form.Get("url"))
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"error": "cant start torrent %s"}`, err)))
		return
	}
}

func handleStopDownload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	torrentClient.Close()
	w.Write([]byte(`{"status": "stopped"}`))

}
