package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	mkv      = ".mkv"
	reString = "((BD|DVD)Rip)|((AAC|FLAC)[0-9]?)|BluRay|HDTV|((X|x)?264-.+)|[0-9]{3,4}p?|mkv|WEB|S[0-9]{2}(E[0-9]{2})?"
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

	if *host == "" || *user == "" || *password == "" {
		flag.Usage()
	}

	pair, err := dialSSHSFTP(*host, &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// remoteFiles := make(map[string]remoteFile)
	remoteFiles := []remoteFile{}

	walker := pair.sftpc.Walk(*dir)
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
		if remoteFilePath == *dir {
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
			for j, fileInfo := range fileInfos {
				rmf[j] = remoteFile{
					title:    filepath.Base(fileInfo.Name()),
					fullPath: filepath.Join(remoteFilePath, fileInfo.Name()),
					FileInfo: fileInfo,
				}
			}
			remoteFiles = append(remoteFiles, remoteFile{
				title:       cleanTitle(remoteFilePath),
				fullPath:    remoteFilePath,
				FileInfo:    walker.Stat(),
				dir:         true,
				remoteFiles: rmf,
			})
			i++
			continue
		}

		if filepath.Ext(remoteFilePath) == mkv {
			remoteFiles = append(remoteFiles, remoteFile{
				title:    cleanTitle(remoteFilePath),
				fullPath: remoteFilePath,
				FileInfo: walker.Stat(),
			})
			i++
		}
	}

	for _, v := range remoteFiles {
		fmt.Println(v.title)
		fmt.Println(v.fullPath)
		if v.dir {
			fmt.Println("dir files:")
			for _, rf := range v.remoteFiles {
				fmt.Println(rf.title)
			}
			fmt.Println("")
		}
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
	title    string
	fullPath string

	dir         bool
	remoteFiles []remoteFile

	os.FileInfo
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
	return strings.Join(parts, " ")
}
