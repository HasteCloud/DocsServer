package main

import (
	"fmt"
	"github.com/micro/go-web"
	"gitlab.com/golang-commonmark/markdown"
	"gopkg.in/flosch/pongo2.v3"
	"net/http"
	"regexp"
)

var template = pongo2.Must(pongo2.FromFile("./assets/index.html"))
var live = true //*flag.Bool("live", false, "whether files are reload live")

func main() {
	service := web.NewService(
		web.Name("web.front"),
	)
	service.Init()

	startServer(service)
	service.Run()
}

func getConfig() GitConfig {
	return GitConfig.New(GitConfig{})
}

func startServer(service web.Service) {
	md := markdown.New(markdown.Quotes("\"\"''"))
	//subnavCache := cache.New(1*time.Hour, 2*time.Hour)
	service.HandleFunc("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		config := getConfig()
		filePath := ""
		if len(config.Icon) > 0 {
			filePath = "./data/source/"+config.Icon
		} else {
			filePath = "./assets/icon.ico"
		}
		http.ServeFile(w, req, filePath)
	})

	service.HandleFunc("/css", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./assets/style.css")
	})

	service.HandleFunc("/git-css", func(w http.ResponseWriter, req *http.Request) {
		config := getConfig()
		http.ServeFile(w, req, "./data/source/"+config.Style)
	})

	service.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		config := getConfig()
		var localCopy = &config
		var filePath = "./data/source/"
		if req.RequestURI == "/" {
			filePath += localCopy.Home
		} else {
			filePath += req.RequestURI
		}

		fileBody := LoadMarkdownFile(filePath)

		body := md.RenderToString([]byte(fileBody))

		if localCopy.GenerateSubNavHeadings {
			//item, found := subnavCache.Get(req.RequestURI)

			if false {
				//localCopy.SubNavItems = item.([]Item)
			} else {
				lowerBound := localCopy.LowerSubNavHeadingBound
				if lowerBound < 1 {
					lowerBound = 1
				}
				upperBound := localCopy.UpperSubNavHeadingBound
				if upperBound > 6 {
					upperBound = 6
				}
				pattern := fmt.Sprintf(`(?m)<h([%d-%d])>(.+)<\/h[%d-%d]>`, lowerBound, upperBound, lowerBound, upperBound)
				re := regexp.MustCompile(pattern)
				headings := re.FindAllStringSubmatch(body, -1)
				localCopy.SubNavItems = make([]Item, len(headings))
				for index, element := range headings {
					localCopy.SubNavItems[index] = Item{
						Path:fmt.Sprintf(`#%d%s`, index, element[2]),
						Level: element[1],
						Title: element[2],
					}
				}
				//subnavCache.Set(req.RequestURI, &localCopy.SubNavItems, cache.DefaultExpiration)
			}
		}


		var err error
		context := pongo2.Context{"body": body, "config": localCopy}

		if live {
			template, err = pongo2.FromFile("./assets/index.html")
		}

		err = template.ExecuteWriter(context, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}