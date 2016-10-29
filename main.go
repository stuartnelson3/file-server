package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/sftp"
	"github.com/stuartnelson3/guac"
	"golang.org/x/crypto/ssh"
)

const (
	mkv      = ".mkv"
	reString = "(?i)^(((BD|DVD)Rip)|((AAC|FLAC)[0-9]?)|BluRay|HDTV|(x?264-.+)|[0-9]{3,4}p?|mkv|WEB|S[0-9]{2}(E[0-9]{2})?)"
)

func main() {
	// Set up a http.FileSystem pointed at a user's home directory on
	// a remote server.
	var (
		host     = flag.String("ssh.host", os.Getenv("FILESERVER_HOST"), "host to connect to via ssh")
		user     = flag.String("ssh.user", os.Getenv("FILESERVER_USER"), "user to connect with via ssh")
		password = flag.String("ssh.password", os.Getenv("FILESERVER_PASSWORD"), "ssh password to connect with via ssh")
		dir      = flag.String("ssh.dir", os.Getenv("FILESERVER_DIR"), "directory to walk on ssh server")
	)
	flag.Parse()

	if *host == "" || *user == "" || *password == "" || *dir == "" {
		flag.Usage()
	}

	// apiMatches := getRemoteFiles(*host, *user, *password, *dir)
	f, err := os.Open("example.json")
	if err != nil {
		log.Fatalf("read: %v", err)
	}
	defer f.Close()

	apiMatches := []remoteFile{}
	if err := json.NewDecoder(f).Decode(&apiMatches); err != nil {
		log.Fatalf("decode: %v", err)
	}

	http.HandleFunc("/api/v0/movies", func(m []remoteFile) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if origin := r.Header.Get("Origin"); origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
				w.Header().Set("Access-Control-Allow-Headers",
					"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			}

			if r.Method == "OPTIONS" {
				return
			}

			json.NewEncoder(w).Encode(m)
		}
	}(apiMatches))

	http.HandleFunc("/script.js", func(w http.ResponseWriter, r *http.Request) {
		script, err := os.Open("script.js")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer script.Close()
		fs, err := script.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, "script", fs.ModTime(), script)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index, err := os.Open("src/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer index.Close()
		fs, err := index.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, "index", fs.ModTime(), index)
	})

	// Recompile the elm code whenever a change is detected.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recompileFn := func() error {
		cmd := exec.Command("elm-make", "src/Main.elm", "--output", "script.js")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	watcher, err := guac.NewWatcher(ctx, "./src", recompileFn)
	if err != nil {
		log.Fatalf("error watching: %v", err)
	}
	go watcher.Run()

	port := "8080"
	log.Printf("starting listener on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}

	// if err := json.NewEncoder(os.Stdout).Encode(remoteFiles); err != nil {
	// 	fmt.Printf("%v", err)
	// }
	// fs, err := sshttp.NewFileSystem(host, sshconfig)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// // Bind HTTP server, provide link for user to browse files
	// host := ":8080"
	// log.Printf("starting listener: http://localhost%s/", host)
	// if err := http.ListenAndServe(":8080", http.FileServer(fs)); err != nil {
	// 	log.Fatal(err)
	// }
}

type remoteFile struct {
	Title    string
	FullPath string

	Dir         bool
	RemoteFiles []remoteFile

	ApiMovie apiMovie

	os.FileInfo
}

type apiResponse struct {
	Search []apiMovie
}

type apiMovie struct {
	Title  string
	Year   string
	ImdbID string
	Type   string
	Poster string
}

type clientPair struct {
	sshc  *ssh.Client
	sftpc *sftp.Client
}

func dialSSHSFTP(host string, config *ssh.ClientConfig) (*clientPair, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "sftp" {
		return nil, fmt.Errorf("invalid URL scheme: %s", u.Scheme)
	}

	// Open initial SSH connection
	sshc, err := ssh.Dial("tcp", u.Host, config)
	if err != nil {
		return nil, err
	}

	// Open SFTP subsystem using SSH connection
	sftpc, err := sftp.NewClient(sshc)
	if err != nil {
		return nil, err
	}

	return &clientPair{
		sshc:  sshc,
		sftpc: sftpc,
	}, nil
}

func cleanTitle(filename string) string {
	base := filepath.Base(filename)
	parts := strings.Split(base, ".")
	for i, part := range parts {
		matched, err := regexp.MatchString(reString, part)
		if err != nil {
			log.Printf("error: %v", err)
		}
		if matched {
			parts[i] = ""
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func getRemoteFiles(host, user, password, dir string) []remoteFile {
	pair, err := dialSSHSFTP(host, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	remoteFiles := []remoteFile{}

	walker := pair.sftpc.Walk(dir)
	i := 0
	for walker.Step() {
		if err := walker.Err(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if i == 10 {
			break
		}
		remoteFilePath := walker.Path()
		if remoteFilePath == dir {
			continue
		}
		// Make worker pool that searches the movie api for the titles

		fi := walker.Stat()
		if fi.IsDir() {
			// We will read the contents and put them in the
			// struct, but we aren't going to return individual
			// files through the API.
			walker.SkipDir()

			fileInfos, err := pair.sftpc.ReadDir(remoteFilePath)
			if err != nil {
				log.Printf("error reading dir %s, skipping\n", remoteFilePath)
				continue
			}

			// Should I check for .mkv in here and ditch the whole dir if it doesn't have a video in it?
			// Probably since that's what I'm doing for the bare files..
			rmf := make([]remoteFile, len(fileInfos), len(fileInfos))
			var videoDir bool
			for j, fileInfo := range fileInfos {
				if filepath.Ext(remoteFilePath) == mkv {
					videoDir = true
				}
				rmf[j] = remoteFile{
					Title:    filepath.Base(fileInfo.Name()),
					FullPath: filepath.Join(remoteFilePath, fileInfo.Name()),
					FileInfo: fileInfo,
				}
			}
			if videoDir {
				remoteFiles = append(remoteFiles, remoteFile{
					Title:       cleanTitle(remoteFilePath),
					FullPath:    remoteFilePath,
					FileInfo:    walker.Stat(),
					Dir:         true,
					RemoteFiles: rmf,
				})
				i++
			}

			continue
		}

		if filepath.Ext(remoteFilePath) == mkv {
			remoteFiles = append(remoteFiles, remoteFile{
				Title:    cleanTitle(remoteFilePath),
				FullPath: remoteFilePath,
				FileInfo: walker.Stat(),
			})
			i++
		}
	}

	urlFormat := "http://www.omdbapi.com/?s=%s"
	for i, v := range remoteFiles {
		// fmt.Println(v.title)
		// fmt.Println(v.fullPath)
		url := fmt.Sprintf(urlFormat, url.QueryEscape(v.Title))
		// fmt.Printf("making request to %s\n", url)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("error fetching movie: %v\n", err)
			continue
		}

		apiResp := apiResponse{}
		json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()

		if len(apiResp.Search) > 0 {
			// fmt.Println(apiResp.Search[0])
			remoteFiles[i].ApiMovie = apiResp.Search[0]
		}
	}

	apiMatches := make([]remoteFile, 0, len(remoteFiles))
	for _, v := range remoteFiles {
		if v.ApiMovie.Poster != "" {
			apiMatches = append(apiMatches, v)
		}
	}
	return apiMatches
}
