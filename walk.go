package main

import (
	"context"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"net/http"
)

type UI struct {
	window   *walk.MainWindow
	btnOpen  *walk.PushButton
	btnClose *walk.PushButton
	path     *walk.LineEdit
	status   *walk.TextLabel
	statusId int
	e        *Engine
	server   *http.Server
}

func (ui *UI) Start() {
	if ui.statusId == 1 {
		return
	}

	ui.statusId = 1
	ui.status.Synchronize(func() {
		ui.status.SetText("状态：运行中")
	})
	ui.e.base = ui.path.Text()

	ui.server = &http.Server{Addr: ui.e.port, Handler: ui.e}
	go ui.server.ListenAndServe()

}

func (ui *UI) Close() {
	if ui.statusId == 0 {
		return
	}

	ui.statusId = 0
	ui.server.Shutdown(context.TODO())

	ui.status.Synchronize(func() {
		ui.status.SetText("状态：已关闭")
	})
}

func createWindow() UI {
	var ui UI
	ui.statusId = 0
	ui.e = &Engine{
		port: ":9999",
	}

	window := declarative.MainWindow{
		AssignTo: &ui.window,
		Title:    "网络文件浏览器",
		MinSize:  declarative.Size{Width: 200, Height: 100},
		Size:     declarative.Size{Width: 200, Height: 200},
		Layout:   declarative.VBox{},
		Children: []declarative.Widget{
			declarative.LineEdit{
				AssignTo:  &ui.path,
				CueBanner: "输入路径",
			},
			declarative.TextLabel{
				AssignTo: &ui.status,
				Text:     "状态：关闭",
			},
			declarative.PushButton{
				AssignTo:  &ui.btnOpen,
				Text:      "开启服务",
				OnClicked: ui.Start,
			},
			declarative.PushButton{
				AssignTo:  &ui.btnClose,
				Text:      "关闭服务",
				OnClicked: ui.Close,
			},
		},
	}

	window.Create()
	return ui
}

func main() {
	ui := createWindow()
	ui.window.Run()
}
