package main

import (
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/app"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func main() {
	// One instance: on Windows and Linux, opening a file from the file
	// manager starts another process, which hands its command line to this
	// one and exits. macOS delivers the file to the running app itself.
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}

	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		log.Fatal(err)
	}

	s, err := settings.Load(dir)
	a := app.New(s, dir)
	if err != nil {
		a.SetError(locales.LoadSettingsFailed() + ": " + err.Error())
		log.Printf("load settings: %v", err)
	}
	if err := settings.Ensure(dir, s); err != nil {
		a.SetError(locales.SaveSettingsFailed() + ": " + err.Error())
		log.Printf("initialize settings: %v", err)
	}
	if err := registerFonts(); err != nil {
		log.Printf("register fonts: %v", err)
		a.SetError(locales.FontLoadFallback())
	}

	// Start in the saved appearance, before the window draws its first
	// frame.
	app.SetThemeSource(s.Theme)

	// Open files the system asks for, from Finder, Explorer or the command
	// line.
	mygo.App.OnOpenFile(a.HandleOpenFile)
	mygo.App.OnSecondInstance(a.HandleSecondInstance)
	mygo.App.OnBeforeQuit(a.BeforeQuit)

	mygo.App.WhenReady(func() {
		a.Prepare()
		mygo.App.SetLocale(s.Language)
		a.RestoreSession()
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "ZPad",
			Width:     1200,
			Height:    800,
			MinWidth:  800,
			MinHeight: 520,
			StateKey:  "main",
			Content:   ui.View(a.View),
		})
		a.Attach(win)
		a.OpenCommandLine(os.Args[1:])
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
