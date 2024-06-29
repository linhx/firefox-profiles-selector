package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"gopkg.in/ini.v1"
)

var url string
var logger *log.Logger

func init() {
	// Initialize the logger
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	if len(os.Args) > 1 {
		url = os.Args[1]
	} else {
		url = "about:newtab"
	}

	gtk.Init(nil)

	win := setupWindow("Firefox profile selector")
	win.ShowAll()
	gtk.Main()
}

const BUTTON_WIDTH = 40

// Create and initialize the window
func setupWindow(title string) *gtk.Window {
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		logger.Fatal("Unable to create window:", err)
	}
	windowWidth := 400
	windowHeight := 200
	win.SetTitle(title)
	win.SetDecorated(false)
	win.SetDefaultSize(windowWidth, windowHeight)
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})
	win.Connect("focus-out-event", func(widget *gtk.Window, event *gdk.Event) bool {
		gtk.MainQuit()
		return false // Continue with the default behavior
	})

	// vertical button box
	mainBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 2)
	if err != nil {
		logger.Fatal("Unable to create mainBox:", err)
	}
	mainBox.SetVAlign(gtk.ALIGN_CENTER)
	mainBox.SetBorderWidth(20)

	// url label
	urlLabel, err := gtk.LabelNew(url)
	if err != nil {
		logger.Fatal("Unable to create btmBox:", err)
	}
	urlLabel.SetMaxWidthChars(250)
	urlLabel.SetEllipsize(pango.ELLIPSIZE_MIDDLE)
	urlLabel.SetMarginBottom(20)
	mainBox.Add(urlLabel)

	// horizontal button box
	btnBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	if err != nil {
		logger.Fatal("Unable to create btmBox:", err)
	}
	btnBox.SetHAlign(gtk.ALIGN_CENTER)

	// get configurations
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	// Get the directory containing the executable
	exeDir := filepath.Dir(exePath)
	cfg, err := ini.Load(filepath.Join(exeDir, "config.ini"))
	if err != nil {
		logger.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	profilesIniPath := cfg.Section("setting").Key("profiles_path").String()
	if strings.HasPrefix(profilesIniPath, "~/") {
		home, _ := os.UserHomeDir()
		profilesIniPath = filepath.Join(home, profilesIniPath[2:])
	}

	// add buttons
	profilesName := getProfiles(profilesIniPath)

	firefoxExecuteFilePath := cfg.Section("setting").Key("exec_path").String()
	if strings.HasPrefix(firefoxExecuteFilePath, "~/") {
		home, _ := os.UserHomeDir()
		firefoxExecuteFilePath = filepath.Join(home, firefoxExecuteFilePath[2:])
	}

	for _, profileName := range profilesName {
		showProfilesButton(btnBox, profileName, firefoxExecuteFilePath)
	}

	mainBox.PackStart(btnBox, false, false, 0)
	win.Add(mainBox)
	return win
}

func getProfiles(profilesIniPath string) []string {
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

func showProfilesButton(buttonContainer *gtk.Box, profileName string, execPath string) {
	btn, _ := gtk.ButtonNew()
	btn.Connect("clicked", func() {
		cmd := exec.Command(execPath, "-P", profileName, "-new-tab", url)
		if err := cmd.Start(); err != nil {
			logger.Printf("Failed to start cmd: %v", err)
			gtk.MainQuit()
			return
		}
		gtk.MainQuit()
		return
	})
	btn.SetLabel(profileName)
	buttonContainer.PackEnd(btn, false, false, 0)
}
