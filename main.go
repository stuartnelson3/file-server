package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	mkv      = ".mkv"
	reString = "((BD|DVD)Rip)|((AAC|FLAC)[0-9]?)|BluRay|HDTV|((X|x)?264-.+)|[0-9]{3}p|mkv|WEB|S[0-9]{2}E[0-9]{2}"
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

	sshconfig := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
	}

	pair, err := dialSSHSFTP(*host, sshconfig)
	if err != nil {
		log.Fatal(err)
	}

	remoteFiles := make(map[string]remoteFile)

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
		// Make worker pool that searches the movie api for the titles
		if filepath.Ext(walker.Path()) == mkv {
			remoteFiles[walker.Path()] = remoteFile{
				title:    filepath.Base(walker.Path()),
				fullPath: walker.Path(),
				FileInfo: walker.Stat(),
			}
			i++
		}
	}
	for _, file := range remoteFiles {
		fmt.Println(file.title)
		fmt.Println(file.fullPath)
	}
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
