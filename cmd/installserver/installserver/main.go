package main

import (
	"io"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	http.HandleFunc("/", DownloadAndServe)
	log.Fatal(http.ListenAndServe(":8087", nil))
}

// DownloadAndServe serve you l15 install script
func DownloadAndServe(w http.ResponseWriter, r *http.Request) {
	version := "main"
	versionQueryParam, ok := r.URL.Query()["version"]

	if ok {
		// trim potential leading 'v'
		versionQueryParam := strings.Replace(versionQueryParam[0], "v", "", 1)

		// https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
		semverRegex := `^(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`

		isVersionString, err := regexp.MatchString(semverRegex, versionQueryParam)

		if nil != err || !isVersionString {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

			return
		}

		if isVersionString {
			version = "v" + versionQueryParam
		}
	}

	resp, err := http.Get("https://raw.githubusercontent.com/lukso-network/network-lukso-cli/" + version + "/shell_scripts/install-unix.sh")

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
