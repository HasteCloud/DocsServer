package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gitlab.com/golang-commonmark/markdown"
	"gopkg.in/flosch/pongo2.v3"
	"gopkg.in/src-d/go-git.v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"github.com/patrickmn/go-cache"
	"time"
)

var config GitConfig
var template = pongo2.Must(pongo2.FromFile("./assets/index.html"))
var live = true //*flag.Bool("live", false, "whether files are reload live")

var (
	port = flag.String("port", "8443", "the port to serve on")
	cert = flag.String("cert", "", "path to certificate file")
	key = flag.String("key", "", "path to key file")
)

func main() {
	flag.Parse()

	checkAndLoadSource()
	startServer()
}

func startServer() {
	md := markdown.New(markdown.Quotes("\"\"''"))
	subnavCache := cache.New(1*time.Hour, 2*time.Hour)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		var filePath = ""
		if len(config.Icon) > 0 {
			filePath = "./data/source/"+config.Icon
		} else {
			filePath = "./assets/icon.ico"
		}
		http.ServeFile(w, req, filePath)
	})

	http.HandleFunc("/css", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./assets/style.css")
	})

	http.HandleFunc("/git-css", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./data/source/"+config.Style)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
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
				subnavCache.Set(req.RequestURI, &localCopy.SubNavItems, cache.DefaultExpiration)
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

	log.Printf("About to listen on %s. Go to http://127.0.0.1:%s/", *port, *port)
	var err error
	if len(*cert) > 0 && len(*key) > 0 {
		err = http.ListenAndServeTLS(":8443", *cert, *key, nil)
	} else {
		err = http.ListenAndServe(":"+*port, nil)
	}
	log.Fatal(err)
}


func checkAndLoadSource() {
	gitURL := os.Getenv("GIT")
	_, err := git.PlainClone("./data/source", false, &git.CloneOptions{
		URL:      gitURL,
		Progress: os.Stdout,
	})

	if err != nil {
		log.Println(err.Error())
		if err == git.ErrRepositoryAlreadyExists {
			log.Println("Pulling from Repo")
			r, err := git.PlainOpen("./data/source")

			tree, err := r.Worktree()
			if err != nil {
				log.Println(err.Error())
				return
			}

			tree.Pull(&git.PullOptions{RemoteName: "origin"})
			log.Println("Pulled")
		} else {
			return
		}
	}

	log.Println("Loading Config")
	LoadConfiguration("./data/source/documentation.json")
	log.Println("Loaded Config")
}

func LoadConfiguration(filePath string) {
	configFile, err := os.Open(filePath)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	config = GitConfig.New(config)
}

func LoadMarkdownFile(filePath string) string {
	if !strings.HasSuffix(filePath, ".md") {
		filePath += ".md"
	}
	return LoadTextFile(filePath)
}

func LoadTextFile(filePath string) string {
	read, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	return string(read)
}