package main

import (
	dd "drurus/drivedb"
	df "drurus/drivefile"
	"strings"
)

var (
	file_index   string = "web/dist/index.html"
	static_index string = "web/dist/"
	// rdb          *redis.Client
)

// AddHostToBase создает и добавляет объекты Host в Redis
func AddHostToBase(host_str string) error {
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
		if err := host.AddRecordDB(); err != nil {
			return err
		}
		// fmt.Println("I'am AddHostToBase: ", host_str)
	}
	return nil
}

func main() {
	defer dd.Rdb.Close()
	// var ctx = context.Background()

	// fmt.Println("iam main")
	// host1 := dd.NewHost("adr1")
	// host2 := dd.NewHost("adr2")

	// host1.IP = "1.1.1.1"
	// host1.IsUse = false
	// host1.Stats = "stat1-None"
	// host1.Tab = "One"

	// host2.IP = "2.2.2.2"
	// host2.IsUse = false
	// host2.Stats = "stat2-None"
	// host2.Tab = "One"

	// hosts := dd.Hosts{}
	// hosts.Add(host1)
	// hosts.Add(host2)

	// fmt.Println("host1", host1)
	// fmt.Println("host12", host2)
	// fmt.Println("hosts", hosts)

	// err := host1.AddRecordDB()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = host2.AddRecordDB()
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// adr3 := dd.NewHost("adr1")
	// _ = adr3.GetRecordDB()
	// fmt.Printf("adr3: %+v\n", adr3)

	// // hosts1 := dd.Hosts{}
	// hosts.Clean()
	// if err := hosts.GetAllHosts(); err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Printf("HOSTS: %+v\n", hosts)

	go df.LoadDirectory("./tabPages", AddHostToBase)

}
