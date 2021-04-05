package main

import (
	dd "drurus/drivedb"
	df "drurus/drivefile"
	dp "drurus/pingtools"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	file_index   string = "web/dist/index.html"
	static_index string = "web/dist/"
)

type WebAnswer struct {
	Data interface{} `json:"data"`
	Err  string      `json:"err"`
}

// AddHostToBase создает и добавляет объекты Host в Redis
//  Вспомогательная функция
func AddHostToBase(host_str, tab string) error {
	host_str = strings.TrimSpace(host_str)

	switch {
	case host_str == "":
		// fmt.Println("SKIP EMPTY LINE")
		return nil
	case strings.HasPrefix(host_str, "#"):
		// fmt.Println("SKIP COMMENT LINE")
		return nil
	default:
		host := dd.NewHost(host_str)
		host.Tab = tab
		if err := host.AddRecordDB(); err != nil {
			return err
		}
		// fmt.Println("I'am AddHostToBase: ", host_str)
	}
	return nil
}

// loadPages по запросу клиента выгружает все хосты из Redis, отдает JSON
func loadPages(c echo.Context) error {
	rsp := &WebAnswer{}
	hs := dd.Hosts{}
	if err := hs.GetAllHosts(); err != nil {
		rsp.Err = err.Error()
	}
	rsp.Data = hs
	return c.JSON(http.StatusOK, rsp)
}

func startWeb() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	e.Static("/", static_index)
	// e.File("/", file_index)
	e.GET("/loadPages", loadPages)
	// e.GET("/getTabPages", getTabPages)
	// e.GET("/readStats", readStats) // get all data from Redis
	// e.GET("/ping/:host", makePingOnce)

	e.Logger.Fatal(e.Start("127.0.0.1:6060"))
}

func startPing() {
	for {
		func() {

			fmt.Println("startPing!")
			hs := dd.Hosts{}
			if err := hs.GetAllHosts(); err != nil {
				fmt.Println(err)
			}

			for _, hst := range hs.Hst {
				fmt.Printf("\t\tping %s\n", hst.Name)
				go func(hst dd.Host) {
					ret, err := dp.ProcessPing(hst.Name)
					if err != nil {
						fmt.Println(hst.Name, err.Error())
					}
					host := dd.NewHost(hst.Name)
					host.IP = ret.IPAddr.String()
					host.Stats = fmt.Sprintf("Loss %.2f%%\nRtt %s",
						ret.PacketLoss,
						ret.AvgRtt.Round(time.Millisecond).String())
					host.Tab = hst.Tab
					if err = host.UpdateRecordDB(); err != nil {
						fmt.Println(hst.Name, err.Error())
					}
				}(hst)
			}
		}()
		<-time.After(time.Second * 15)
	}
}

func main() {
	defer dd.Rdb.Close()
	// var ctx = context.Background()

	for {
		err := df.LoadDirectory("./tabPages", AddHostToBase)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 3)
		} else {
			break
		}
	}

	go startPing()
	startWeb()
}
