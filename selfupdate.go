package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"

	"golang.org/x/mod/semver"

	"github.com/valyala/fastjson"

	"github.com/minio/selfupdate"

	"github.com/ulikunitz/xz"
)

func selfUpdate() {

	selfupdate_options := selfupdate.Options{}

	resp, err := http.Get("https://api.github.com/repos/babs/multiping/releases/latest")
	if err != nil {
		log.Fatalln(err)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(payload)
	if err != nil {
		log.Fatalln(err)
	}

	latest_release := string(v.GetStringBytes("name"))

	log.Printf("current version: %v, latest release: %v", Version, latest_release)

	switch semver.Compare(latest_release, Version) {
	case -1:
		log.Println("you have a newer version, wut ?!?")
		return
	case 0:
		log.Println("already latest version")
		return
	case 1:
		log.Println("new version detected, upgrading")
		if Version == "v0.0.0" {
			log.Println("development release detected, press enter to proceed")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
	}

	ext := "xz"
	if runtime.GOOS == "windows" {
		ext = "exe.xz"
	}

	download_link := fmt.Sprintf("https://github.com/babs/multiping/releases/download/%v/multiping-%v-%v.%v", latest_release, runtime.GOOS, runtime.GOARCH, ext)

	err = selfupdate_options.CheckPermissions()
	if err != nil {
		log.Printf("won't perform self update: %s\n\nIf you want to update manually, you can get the latest version for your platform at:\n -> %s\n", err, download_link)

		return
	}

	log.Printf("downloading %v\n", download_link)

	resp, err = http.Get(download_link)
	if err != nil {
		log.Fatalln(err)
	}
	r, err := xz.NewReader(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	err = selfupdate.Apply(r, selfupdate_options)
	if err == nil {
		log.Println("upgrade complete, happy mutliping-ing :-)")
	} else {
		log.Fatalf("unable to upgrade: %v", err)
	}
}
