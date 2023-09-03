package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

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

func createNamedPipe(pipePath string) error {
	err := syscall.Mkfifo(pipePath, 0666)
	if err != nil {
		return err
	}
	return nil
}

func writeToPipe(pipePath string, message string) error {
	file, err := os.OpenFile(pipePath, os.O_WRONLY, os.ModeNamedPipe)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(message)
	if err != nil {
		return err
	}

	return nil
}

const (
	lockFilePath = "/tmp/firefox-profiles-selector.lock"
	pipePath     = "/tmp/firefox-profiles-selector.pipe"
)

func main() {
	if len(os.Args) > 1 {
		url = os.Args[1]
	} else {
		url = "about:newtab"
	}

	_, err := os.Stat(lockFilePath)
	if err == nil {
		writeToPipe(pipePath, url)
		os.Exit(0)
	} else {
		createNamedPipe(pipePath)
		lockFile, err := os.Create(lockFilePath)
		if err != nil {
			fmt.Println("Failed to create the lock file:", err)
			os.Exit(1)
		}
		defer lockFile.Close()

		gtk.Init(nil)

		win := setupWindow("Firefox profile selector")

		win.Connect("destroy", func() {
			if err := os.Remove(lockFilePath); err != nil {
				fmt.Println("Failed to remove the lock file:", err)
			}
		})

		win.ShowAll()
		gtk.Main()
		file, err := os.OpenFile(pipePath, os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			fmt.Println("Failed to open the pipe:", err)
			os.Exit(1)
		}
		defer file.Close()

		// Listen for commands from the CLI app
		for {
			buffer := make([]byte, 1024)
			n, err := file.Read(buffer)
			if err != nil {
				fmt.Println("Error reading from the pipe:", err)
				continue
			}

			command := string(buffer[:n])
			fmt.Println("Received command from CLI:", command)
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
