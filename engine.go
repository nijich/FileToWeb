package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Engine struct {
	port     string
	base     string
	username string
	password string
	auth     bool
}

func (e Engine) handleIcon(w http.ResponseWriter, r *http.Request) {}

func checkVideoExt(ext string) bool {
	if len(ext) < 1 {
		return false
	}
	check := false
	vidSuffix := []string{"avi", "mpg", "mpe", "mpeg", "asf", "wmv", "mov", "qt", "rm", "mp4", "flv", "m4v", "webm", "ogv", "ogg", "mkv", "ts", "tsv"}

	for _, val := range vidSuffix {
		if ext[1:] == val {
			check = true
			break
		}
	}
	return check
}

func (e Engine) Serve(w http.ResponseWriter, r *http.Request) {

	if e.auth {
		username, password, _ := r.BasicAuth()

		if username != e.username || password != e.password {
			w.Header().Set("WWW-Authenticate", "Basic realm==")
			w.WriteHeader(401)
			w.Write([]byte("认证失败"))
			return
		}
	}

	link := r.URL.Path
	local := e.base + strings.Replace(link, "/", "\\", -1)
	filename := filepath.Base(local)

	//log.Println(e.base)

	if !strings.HasSuffix(link, "/") {
		t, err := os.Open(local)
		if err != nil {
			log.Println(err)
		}

		ext := filepath.Ext(local)
		if checkVideoExt(ext) {
			w.Header().Add("Content-Type", "video")
		}
		http.ServeContent(w, r, filename, time.Now(), t)
	} else {
		files, _ := ioutil.ReadDir(local)
		for _, file := range files {
			link := link + file.Name()
			if file.IsDir() {
				link += "/"
			}
			w.Write([]byte("<a href=\"" + link + "\">" + file.Name() + "</a><br/>"))
		}
	}
}

func (e Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/favicon.ico":
	default:
		e.Serve(w, r)
	}
}
