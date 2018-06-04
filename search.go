package main

import (
	"net/http"
	"net/url"
	"sync"

	"astuart.co/goq"
)

//TorrentInfo torrent information
type TorrentInfo struct {
	Name       string `json:"name"`
	TorrentURL string `json:"url"`
	MagnetLink string `json:"magnet"`
	Seeds      string `json:"seed"`
	Leeches    string `json:"leech"`
	Size       string `json:"size"`
}

type rutorResults struct {
	Files []struct {
		Data  []string `goquery:"td"`
		Links []string `goquery:"a,[href]"`
		Seed  string   `goquery:"span.green,text"`
		Leech string   `goquery:"span.red,text"`
	} `goquery:"#index table tbody tr,html"`
}

func searchRutor(term string) (torrents []TorrentInfo) {
	var wg sync.WaitGroup
	var m sync.Mutex
	cats := []string{"1", "5", "7"} // "5" - foreign, "1" - russian, "7" - cartoons
	for _, cat := range cats {
		wg.Add(1)
		go func(cat string) {
			defer wg.Done()
			if results, err := searchRutorCategory(cat, term); err == nil {
				m.Lock()
				torrents = append(torrents, results...)
				m.Unlock()
			}
		}(cat)
	}
	wg.Wait()
	return
}

func searchRutorCategory(cat string, term string) (torrents []TorrentInfo, err error) {
	addr, err := url.Parse("http://rutor.is/search/0/" + cat + "/100/2/" + term)
	if err != nil {
		return
	}
	res, err := http.Get(addr.String())

	if err != nil {
		return
	}
	defer res.Body.Close()

	var r rutorResults

	err = goq.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return
	}
	for _, t := range r.Files {
		torrent := TorrentInfo{
			Seeds:   t.Seed,
			Leeches: t.Leech,
		}
		if len(t.Data) > 3 {
			torrent.Name = t.Data[1]
			torrent.Size = t.Data[3]
		}
		if len(t.Links) > 1 {
			torrent.TorrentURL = "http://rutor.is" + t.Links[0]
			torrent.MagnetLink = t.Links[1]

			torrents = append(torrents, torrent)
		}
	}
	return
}
