package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// TODO: consider to move it to common/shared
const (
	pandoraDependencyName   = "pandora"
	vanguardDependencyName  = "vanguard"
	validatorDependencyName = "validator"
)

var (
	clientDependencies = map[string]*ClientDependency{
		pandoraDependencyName: {
			baseUnixUrl: "https://github.com/lukso-network/pandora-execution-engine/releases/download/%s/geth",
			name:        pandoraDependencyName,
		},
		vanguardDependencyName: {
			baseUnixUrl: "https://github.com/lukso-network/vanguard-consensus-engine/releases/download/%s/beacon-chain",
			name:        vanguardDependencyName,
		},
		validatorDependencyName: {
			baseUnixUrl: "https://github.com/lukso-network/vanguard-consensus-engine/releases/download/%s/validator",
			name:        validatorDependencyName,
		},
	}
)

type ClientDependency struct {
	baseUnixUrl string
	name        string
}

func (dependency *ClientDependency) ParseUrl(tagName string) (url string) {
	return fmt.Sprintf(dependency.baseUnixUrl, tagName)
}

func (dependency *ClientDependency) ResolveDirPath(tagName string, datadir string) (location string) {
	location = fmt.Sprintf("%s/%s", datadir, tagName)

	return
}

func (dependency *ClientDependency) ResolveBinaryPath(tagName string, datadir string) (location string) {
	location = fmt.Sprintf("%s/%s", dependency.ResolveDirPath(tagName, datadir), dependency.name)

	return
}

func (dependency *ClientDependency) Download(tagName string, destination string) (err error) {
	dependencyTagPath := dependency.ResolveDirPath(tagName, destination)
	err = os.MkdirAll(dependencyTagPath, 0755)

	if nil != err {
		return
	}

	dependencyLocation := dependency.ResolveBinaryPath(tagName, destination)

	if fileExists(dependencyLocation) {
		log.Warning("I am not downloading pandora, file already exists")

		return
	}

	fileUrl := dependency.ParseUrl(tagName)
	response, err := http.Get(fileUrl)

	if nil != err {
		return
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if http.StatusOK != response.StatusCode {
		return fmt.Errorf(
			"invalid response when downloading pandora on file url: %s. Response: %s",
			fileUrl,
			response.Status,
		)
	}

	output, err := os.Create(dependencyLocation)

	if nil != err {
		return
	}

	defer func() {
		_ = output.Close()
	}()

	_, err = io.Copy(output, response.Body)

	if nil != err {
		return
	}

	err = os.Chmod(dependencyLocation, os.ModePerm)

	return
}
