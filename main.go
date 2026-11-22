package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/helloxz/zpad/internal/app"
	"github.com/helloxz/zpad/internal/settings"
)

func main() {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		log.Fatal(err)
	}

	s, err := settings.Load(dir)
	a := app.New(s, dir)
	if err != nil {
		a.SetError("读取设置失败: " + err.Error())
		log.Printf("load settings: %v", err)
	}
	if err := settings.Ensure(dir, s); err != nil {
		a.SetError("保存设置失败: " + err.Error())
		log.Printf("initialize settings: %v", err)
	}

	// Start in the saved appearance, before the window draws its first
	// frame.
	app.SetThemeSource(s.Theme)

	// Open files the system asks for, from Finder or the command line.
	mygo.App.OnOpenFile(a.HandleOpenFile)
	mygo.App.OnBeforeQuit(a.BeforeQuit)

	mygo.App.WhenReady(func() {
		a.Prepare()
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
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
