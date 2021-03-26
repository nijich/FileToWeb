package main

import (
	"context"
	"fmt"
	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"
	"github.com/skip2/go-qrcode"
	"net"
	"net/http"
	"strconv"
)

type UI struct {
	window   *walk.MainWindow
	btnStart *walk.PushButton
	btnClose *walk.PushButton
	path     *walk.LineEdit
	port     *walk.LineEdit
	status   *walk.TextLabel
	qrcode   *walk.ImageView
	addrs    *walk.ComboBox
	auth     *walk.CheckBox
	username *walk.LineEdit
	password *walk.LineEdit
	e        *Engine
	server   *http.Server
}

func (ui *UI) checkPort() (e string) {
	port := ui.port.Text()
	num, err := strconv.Atoi(port)
	if err != nil || num < 1 || num > 65535 {
		e = "无效的端口号"
	} else {
		e = ""
		ui.e.port = ":" + port
	}
	return e
}

func (ui *UI) Start() {

	errString := ui.checkPort()
	if errString != "" {
		ui.status.Synchronize(func() {
			ui.status.SetText("错误：" + errString)
		})
		return
	}

	ui.RefreshQRCode()

	ui.status.SetText("状态：运行中")
	ui.btnStart.SetEnabled(false)
	ui.btnClose.SetEnabled(true)

	if ui.auth.Enabled() {
		ui.e.auth = true
		ui.e.username = ui.username.Text()
		ui.e.password = ui.password.Text()
	} else {
		ui.e.auth = false
	}

	ui.e.base = ui.path.Text()

	ui.server = &http.Server{Addr: ui.e.port, Handler: ui.e}
	go ui.server.ListenAndServe()
}

func (ui *UI) Close() {
	ui.btnStart.SetEnabled(true)
	ui.btnClose.SetEnabled(false)

	ui.server.Shutdown(context.TODO())

	ui.status.Synchronize(func() {
		ui.status.SetText("状态：已关闭")
	})
}

func createWindow() UI {
	var ui UI

	ui.e = &Engine{
		port: ":9999",
		auth: false,
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
				CueBanner: "输入使用的端口号",
				//OnEditingFinished: ui.RefreshQRCode,
			},
			declarative.PushButton{
				AssignTo:  &ui.btnStart,
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
			declarative.CheckBox{
				AssignTo: &ui.auth,
				Text:     "启用密码验证",
			},
			declarative.LineEdit{
				AssignTo:  &ui.username,
				CueBanner: "默认为空",
			},
			declarative.LineEdit{
				AssignTo:  &ui.password,
				CueBanner: "默认为空",
			},
		},
	}
	window.Create()
	return ui
}

func (ui *UI) RefreshQRCode() {
	ip := ui.addrs.Text()
	ui.qrcode.SetImage(genQRCode(ip, ui.e.port))
}

func (ui *UI) initInfo() {

	addrs := getLocalIP()
	ui.addrs.SetModel(addrs)
	ui.addrs.SetCurrentIndex(0)
	ui.btnClose.SetEnabled(false)
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
	qr, err := qrcode.New("http://"+ip+port, qrcode.Medium)
	if err != nil {
		return nil
	}
	image, _ := walk.NewBitmapFromImageForDPI(qr.Image(1000), 1000)
	return image
}
