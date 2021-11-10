package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/", DownloadAndServe)
	log.Fatal(http.ListenAndServe(":8087", nil))
}

// DownloadAndServe serve you l15 install script
func DownloadAndServe(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://raw.githubusercontent.com/lukso-network/network-lukso-cli/main/shell_scripts/install-unix.sh")

	if nil != err {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	w.Header().Set("Content-Disposition", "attachment; filename=lukso.sh")
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	_, err = io.Copy(w, resp.Body)

	if nil != err {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}
}
