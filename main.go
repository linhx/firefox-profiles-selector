package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"gopkg.in/ini.v1"
)

var timeoutContinue = true
var tos glib.SourceHandle

// Create and initialize the window
func setupWindow(title string) *gtk.Window {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetTitle(title)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.SetPosition(gtk.WIN_POS_CENTER)
	width, height := 200, 100
	win.SetDefaultSize(width, height)

	box, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 2)

	profilesName := getProfiles()

	for _, profileName := range profilesName {
		showProfilesButton(box, profileName)
	}

	win.Add(box)

	return win
}

var url string

const (
	socketPath = "/tmp/firefox-profiles-selector.sock"
)

var isDestroyed = false

func main() {
	if len(os.Args) > 1 {
		url = os.Args[1]
	} else {
		url = "about:newtab"
	}

	_, err := os.Stat(socketPath)
	if err != nil {
		os.Remove(socketPath)
		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			fmt.Println("Error creating listener:", err)
			return
		}
		defer listener.Close()
		go func() {
			for {
				// Accept connections from clients
				conn, err := listener.Accept()
				if err != nil {
					fmt.Println("Error accepting connection:", err)
					continue
				}
				defer conn.Close()

				// Handle incoming data
				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				if err != nil {
					fmt.Println("Error reading from client:", err)
					continue
				}

				// Process the received data
				data := buffer[:n]
				url = string(data)
			}
		}()
		gtk.Init(nil)
		win := setupWindow("Firefox profile selector")

		win.ShowAll()
		gtk.Main()
	} else {
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer conn.Close()
		// Send data to the server
		_, err = conn.Write([]byte(url))
		if err != nil {
			fmt.Println("Error writing to server:", err)
			return
		}
	}
}

func getProfiles() []string {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	profilesIniPath := cfg.Section("setting").Key("profiles_path").String()
	if strings.HasPrefix(profilesIniPath, "~/") {
		home, _ := os.UserHomeDir()
		profilesIniPath = filepath.Join(home, profilesIniPath[2:])
	}

	profilesCfg, err := ini.Load(profilesIniPath)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	profilesSecs := profilesCfg.Sections()
	profilesName := []string{}
	for _, v := range profilesSecs {
		if strings.HasPrefix(v.Name(), "Profile") {
			profilesName = append(profilesName, v.Key("Name").String())
		}
	}
	return profilesName
}

func showProfilesButton(buttonContainer *gtk.Box, profileName string) {
	btn, _ := gtk.ButtonNew()
	btn.Connect("clicked", func() {
		cmd := exec.Command("firefox", "-P", profileName, "-new-tab", url)
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start cmd: %v", err)
			return
		}
	})
	btn.SetLabel(profileName)
	buttonContainer.Add(btn)
}
