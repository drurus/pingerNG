package main

import (
	"drurus/config"
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
	worker_count int    //= 0 //count
	delay_global int    //= 0 // seconds
	separator    string
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
	// сделать подмену Stats (вместо списка ключей выдать значения)
	// 1) итерировать список хостов
	// 2) получить список стат-ключей
	// 3) загрузить из БД значения по стат-ключам, итерировать
	// 4) заменить host.Stats структурами с данными
	// oStat := make([]string, 0)
	skeys := strings.Split(dd.Sk_template, ",")
	for i, host := range hs.Hst {
		oStat := make([]dd.Stat, 0)
		// перебор стат-ключей
		for _, skey := range skeys {
			// !! сформировать имя полного стат-ключа !!
			realkey := host.Name + separator + skey
			sk := dd.Stat{Type: skey, Values: host.Rsk(realkey)}
			oStat = append(oStat, sk)
			// fmt.Printf("For %s prep %v\n", host.Name, sk)
		}

		hs.Hst[i].Stats = oStat
	}
	rsp.Data = hs
	// fmt.Printf("%+v", rsp)
	return c.JSON(http.StatusOK, rsp)
}

func startWeb() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	e.Static("/", static_index)
	// e.File("/", file_index)
	go e.GET("/loadPages", loadPages)

	s := &http.Server{
		Addr:         "127.0.0.1:6060",
		ReadTimeout:  20 * time.Minute,
		WriteTimeout: 20 * time.Minute,
	}
	e.Debug = true
	e.Logger.Fatal(e.StartServer(s))
	// e.Logger.Fatal(e.Start("127.0.0.1:6060"))
}

// workerPing процесс осуществляющий пинг
func workerPing(id int, jobs <-chan dd.Host) {
	// обработка паники (например не получен IP адрес)
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Printf("wr %d: %s\n", id, err)
	// 		// fmt.Printf("wr %d: %s %s\n", id, hst.Name, err)
	// 	}
	// }()

	for hst := range jobs {
		// fmt.Printf("wr %d <- %s \n", id, hst.Name)
		ret, err := dp.ProcessPing(hst.Name)
		if err != nil {
			fmt.Printf("wr %d: %s %s\n", id, hst.Name, err.Error())
			// continue // чтобы не дойти до паники
		}
		host := dd.NewHost(hst.Name)
		host.IP = ret.IPAddr.String()
		// TODO new stats - набор ключей брать из конфига/web
		// host.Stats = "rtt,loss"
		host.Stats = []dd.Stat{}

		skeys := strings.Split(dd.Sk_template, ",")
		if len(skeys) > 0 {
			for _, key := range skeys {
				values := []string{}
				switch key {
				case "rtt":
					values = append(values, fmt.Sprintf("%d", ret.AvgRtt.Round(time.Millisecond).Milliseconds()))
				case "loss":
					values = append(values, fmt.Sprintf("%.2f", ret.PacketLoss))
				default:
				}
				// fmt.Println("------ VALUES", values)
				if values == nil {
					fmt.Println("-------NIIL")
				}
				host.SaveStats(key, values)
			}
		}

		host.Tab = hst.Tab
		if err = host.UpdateRecordDB(); err != nil {
			fmt.Printf("wr %d: %s %s\n", id, hst.Name, err.Error())
		}
	}
}

// startJobs запускает циклическое чтение хостов и передачу в канал заданий
func startJobs(jobs chan dd.Host) {
	// for {
	// fmt.Printf("\n **************** Start a new cycle of jobs! ****************\n")
	hs := dd.Hosts{}
	if err := hs.GetAllHosts(); err != nil {
		fmt.Println(err, " список загружен не полностью!")
		// continue
	}

	for {
		fmt.Printf("\n **************** Start a new cycle of jobs! ****************\n")
		for _, hst := range hs.Hst {
			// fmt.Printf("%s -> jobs\n", hst.Name)
			jobs <- hst
		}
		// fmt.Printf("----- Global delay %d second -----\n", delay_global)
		<-time.After(time.Duration(delay_global) * time.Second)
	}
}

func main() {
	defer dd.Rdb.Close()

	cnf, err := config.ConfigLoad()
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	fmt.Println(cnf)
	worker_count = int(cnf.WorkerCount)
	delay_global = int(cnf.DelayGlobal)
	separator = cnf.Separator
	fmt.Println(worker_count, delay_global)

	// var ctx = context.Background()

	// крутит цикл чтения пока не заработает Redis
	for {
		err := df.LoadDirectory("./tabPages", AddHostToBase)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 3)
		} else {
			break
		}
	}

	jobs := make(chan dd.Host, worker_count)

	// создать пул воркеров
	for w := 1; w <= worker_count; w++ {
		go workerPing(w, jobs)
	}

	// запустить чтение хостов и создание заданий
	go startJobs(jobs)
	startWeb()
}
