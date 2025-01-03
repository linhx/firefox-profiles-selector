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
	if len(os.Args) > 1 && len(os.Args[1]) > 2 {
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
	// quit on out blur
	win.Connect("focus-out-event", func(widget *gtk.Window, event *gdk.Event) bool {
		gtk.MainQuit()
		return false // Continue with the default behavior
	})
	// quit on press ESC
	win.Connect("key-press-event", func(window *gtk.Window, event *gdk.Event) {
		keyEvent := &gdk.EventKey{Event: event}
		if keyEvent.KeyVal() == gdk.KEY_Escape {
			gtk.MainQuit()
		}
	})

	// vertical button box
	mainBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 2)
	if err != nil {
		logger.Fatal("Unable to create mainBox:", err)
	}
	mainBox.SetVAlign(gtk.ALIGN_CENTER)
	mainBox.SetBorderWidth(20)

	// add url view
	urlView := createUrlView(url)
	mainBox.Add(urlView)

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

func createUrlView(_url string) *gtk.ScrolledWindow {
	// Create a ScrolledWindow
	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("Unable to create scrolled window:", err)
	}

	// Set scrolled window properties
	scrolledWindow.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)

	// Create a TextView
	textView, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal("Unable to create text view:", err)
	}

	// Set TextView as read-only
	textView.SetEditable(false)
	textView.SetCursorVisible(false)
	textView.SetWrapMode(gtk.WRAP_CHAR)

	// Get the TextView buffer and set some text
	buffer, err := textView.GetBuffer()
	if err != nil {
		log.Fatal("Unable to get text buffer:", err)
	}
	buffer.SetText(_url)

	// Add the TextView to the ScrolledWindow
	scrolledWindow.Add(textView)
	scrolledWindow.SetMarginBottom(20)
	return scrolledWindow
}
