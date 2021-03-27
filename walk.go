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
	"strings"
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

func (ui *UI) check() (e string) {
	port := ui.port.Text()
	num, err := strconv.Atoi(port)
	if strings.Trim(ui.path.Text(), " ") == "" {
		e = "无效的路径"
	} else if err != nil || num < 1 || num > 65535 {
		e = "无效的端口号"
	} else {
		e = ""
		ui.e.port = ":" + port
	}
	return e
}

func (ui *UI) Start() {

	errString := ui.check()
	if errString != "" {
		ui.status.SetText("错误：" + errString)
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
	var err error
	go func() {
		err = ui.server.ListenAndServe()
	}()
	if err != nil {
		ui.status.SetText("错误：启动服务器失败")
		ui.btnStart.SetEnabled(true)
		ui.btnClose.SetEnabled(false)
	}
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
	icon, _ := walk.NewIconFromFile("app.ico")

	window := declarative.MainWindow{
		AssignTo: &ui.window,
		Title:    "网络文件浏览器",
		Icon:     icon,
		MinSize:  declarative.Size{Width: 600, Height: 400},
		Size:     declarative.Size{Width: 600, Height: 400},
		//Bounds: declarative.Rectangle{
		//	X: 600,
		//	Y: 200,
		//},
		Layout: declarative.VBox{},
		Children: []declarative.Widget{

			declarative.Composite{
				Layout: declarative.Grid{},
				Children: []declarative.Widget{
					declarative.Composite{
						Row:    0,
						Column: 0,
						Layout: declarative.Grid{},
						Children: []declarative.Widget{
							declarative.TextLabel{
								Text:   "本地路径:",
								Row:    0,
								Column: 0,
							},
							declarative.LineEdit{
								AssignTo:  &ui.path,
								CueBanner: "例如: E:\\",
								Row:       0,
								Column:    1,
							},
							declarative.TextLabel{
								Text:   "端口号:",
								Row:    1,
								Column: 0,
							},
							declarative.LineEdit{
								AssignTo:  &ui.port,
								CueBanner: "例如: 9999",
								Row:       1,
								Column:    1,
								//OnEditingFinished: ui.RefreshQRCode,
							},

							declarative.CheckBox{
								AssignTo: &ui.auth,
								Text:     "启用密码验证",
								Row:      2,
								Column:   0,
							},
							declarative.TextLabel{
								Text:   "用户名:",
								Row:    3,
								Column: 0,
							},
							declarative.LineEdit{
								AssignTo:  &ui.username,
								CueBanner: "默认为空",
								Row:       3,
								Column:    1,
							},
							declarative.TextLabel{
								Text:   "密码:",
								Row:    4,
								Column: 0,
							},
							declarative.LineEdit{
								AssignTo:  &ui.password,
								CueBanner: "默认为空",
								Row:       4,
								Column:    1,
							},

							declarative.Composite{
								Row:        5,
								Column:     0,
								ColumnSpan: 2,
								Layout:     declarative.HBox{},
								Children: []declarative.Widget{

									declarative.PushButton{
										AssignTo:  &ui.btnStart,
										Text:      "开启服务并生成二维码",
										OnClicked: ui.Start,
										MinSize:   declarative.Size{150, 10},
									},
									declarative.HSpacer{},
									declarative.PushButton{
										AssignTo:  &ui.btnClose,
										Text:      "关闭服务",
										OnClicked: ui.Close,
										MinSize:   declarative.Size{150, 10},
									},
								},
							},
						},
					},
					declarative.Composite{
						Background: declarative.SolidColorBrush{walk.RGB(255, 255, 255)},
						Row:        0,
						Column:     1,
						Layout:     declarative.VBox{},
						Children: []declarative.Widget{
							declarative.Label{
								Text: "二维码使用的本地地址:",
							},
							declarative.ComboBox{
								AssignTo: &ui.addrs,
								//OnCurrentIndexChanged: ui.RefreshQRCode, // seems invoke three times once
							},
							declarative.ImageView{
								MinSize:  declarative.Size{200, 200},
								AssignTo: &ui.qrcode,
							},
						},
					},

					declarative.TextLabel{
						AssignTo:      &ui.status,
						Text:          "状态：关闭",
						Row:           1,
						Column:        0,
						ColumnSpan:    2,
						TextAlignment: declarative.AlignHCenterVCenter,
					},
				},
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
	ui.qrcode.SetImage(genQRCode(" ", ""))
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
	qr, err := qrcode.New("http://"+ip+port, qrcode.High)
	if err != nil {
		return nil
	}
	image, _ := walk.NewBitmapFromImageForDPI(qr.Image(1000), 500)
	return image
}
