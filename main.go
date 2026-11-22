package main

import (
	"log"
	"os"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/app"
	"github.com/helloxz/zpad/internal/settings"
	"github.com/helloxz/zpad/locales"
)

func main() {
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

	// Open files the system asks for, from Finder or the command line.
	mygo.App.OnOpenFile(a.HandleOpenFile)
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
		var args []string
		for _, arg := range os.Args[1:] {
			if strings.HasPrefix(arg, "-") {
				continue
			}
			args = append(args, arg)
		}
		if len(args) > 0 {
			a.OpenPaths(args)
		}
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
