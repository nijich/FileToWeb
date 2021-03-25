package main

import (
	"context"
	"fmt"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"github.com/skip2/go-qrcode"
	"net"
	"net/http"
)

type UI struct {
	window   *walk.MainWindow
	btnOpen  *walk.PushButton
	btnClose *walk.PushButton
	path     *walk.LineEdit
	port     *walk.LineEdit
	status   *walk.TextLabel
	qrcode   *walk.ImageView
	addrs    *walk.ComboBox
	statusId int
	e        *Engine
	server   *http.Server
}

func (ui *UI) Start() {
	if ui.statusId == 1 {
		return
	}

	ui.statusId = 1

	ui.e.port = ":" + ui.port.Text()

	ui.RefreshQRCode()

	ui.status.Synchronize(func() {
		ui.status.SetText("状态：运行中")
	})
	ui.port.Synchronize(func() {
		ui.port.SetEnabled(false)
	})
	ui.path.Synchronize(func() {
		ui.path.SetEnabled(false)
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

	ui.port.Synchronize(func() {
		ui.port.SetEnabled(true)
	})
	ui.path.Synchronize(func() {
		ui.path.SetEnabled(true)
	})

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
			declarative.TextLabel{
				AssignTo: &ui.status,
				Text:     "状态：关闭",
			},
			declarative.LineEdit{
				AssignTo:  &ui.path,
				CueBanner: "输入路径",
			},
			declarative.LineEdit{
				AssignTo:  &ui.port,
				CueBanner: "输入使用的端口号，默认:9999",
				//OnEditingFinished: ui.RefreshQRCode,
			},
			declarative.PushButton{
				AssignTo:  &ui.btnOpen,
				Text:      "开启服务并生成二维码",
				OnClicked: ui.Start,
			},
			declarative.PushButton{
				AssignTo:  &ui.btnClose,
				Text:      "关闭服务",
				OnClicked: ui.Close,
			},
			declarative.ComboBox{
				AssignTo: &ui.addrs,
				//OnCurrentIndexChanged: ui.RefreshQRCode, // seems invoke three times once
			},
			declarative.ImageView{
				AssignTo: &ui.qrcode,
			},
		},
	}
	window.Create()
	return ui
}

func (ui *UI) RefreshQRCode() {
	ip := ui.addrs.Text()
	ui.qrcode.Synchronize(
		func() {
			ui.qrcode.SetImage(
				genQRCode(ip, ui.e.port))
		})
}

func (ui *UI) initInfo() {
	addrs := getLocalIP()

	ui.addrs.Synchronize(func() {
		ui.addrs.SetModel(addrs)
		ui.addrs.SetCurrentIndex(0)
	})
}

func main() {

	ui := createWindow()
	ui.initInfo()
	ui.window.Run()
	//x := getLocalIP()
	//fmt.Println(x)
	//genQRCode(x[0], ":9999", "qrcode")
}

func getLocalIP() (result []string) {
	conn, err := net.Dial("ip:icmp", "baidu.com")
	validAddr := ""
	if err == nil {
		validAddr = conn.LocalAddr().String()
		result = append(result, validAddr)
	}

	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
	}

	for _, addr := range addrs {

		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil && ipnet.IP.String() != validAddr {
				result = append(result, ipnet.IP.String())
			}
		}
	}
	return result
}

func genQRCode(ip string, port string) walk.Image {
	qr, err := qrcode.New(ip+port, qrcode.Medium)
	if err != nil {
		return nil
	}
	image, _ := walk.NewBitmapFromImageForDPI(qr.Image(1000), 1000)
	return image
}
