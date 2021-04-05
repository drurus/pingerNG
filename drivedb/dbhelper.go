package drivedb

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	ctx = context.Background()
	Rdb *redis.Client
)

const (
	poolsize = 100
)

func init() {
	Rdb = InitRedis(ctx)
}

type Host struct {
	Name  string `json:"name"`  // hostname
	IP    string `json:"ip"`    // адрес
	IsUse bool   `json:"isuse"` // флаг обработки, 0 - false
	Stats string `json:"stats"` // результат обработки
	Tab   string `json:"tab"`   // имя таба
}

type Hosts struct {
	Hst []Host
}

// type Hosts struct {
// 	Data []Host `json:"data"`
// 	Err  string `json:"err"`
// }

// ToStrings выводит структуру Host как []string
func (h *Host) ToStrings() []string {
	sc := make([]string, 0)
	sc = append(sc, "Name", h.Name)
	sc = append(sc, "IP", h.IP)
	sc = append(sc, "IsUse", strconv.FormatBool((h.IsUse)))
	sc = append(sc, "Stats", h.Stats)
	sc = append(sc, "Tab", h.Tab)
	return sc
}

// GetRecordDB получает запись хоста из Redis по атрибуту Name
func (h *Host) GetRecordDB() error {
	hh, err := Rdb.HGetAll(ctx, (*h).Name).Result()
	if err != nil {
		return err
	}

	(*h).Name = hh["Name"]
	(*h).IP = hh["IP"]
	(*h).IsUse, err = strconv.ParseBool(hh["IsUse"])
	(*h).Stats = hh["Stats"]
	(*h).Tab = hh["Tab"]
	if err != nil {
		return err
	}
	return nil
}

// AddRecordDB добавляет запись хоста в Redis
func (h *Host) AddRecordDB() error {
	r := Rdb.HExists(ctx, (*h).Name, "Name")
	if r.Err() != nil {
		return r.Err()
	}
	if b, err := r.Result(); err != nil {
		return r.Err()
	} else {
		if !b {
			r := Rdb.HSet(ctx, (*h).Name, h.ToStrings())
			if r.Err() != nil {
				return r.Err()
			}
		}
	}
	return nil
}

func (h *Host) UpdateRecordDB() error {
	var mtx sync.RWMutex
	r := Rdb.HExists(ctx, (*h).Name, "Name")
	if r.Err() != nil {
		return r.Err()
	}
	if b, err := r.Result(); err != nil {
		return r.Err()
	} else {
		if b {
			mtx.RLock()
			r := Rdb.HSet(ctx, (*h).Name, h.ToStrings())
			mtx.RUnlock()
			if r.Err() != nil {
				return r.Err()
			}
		}
	}
	return nil
}

func (hs *Hosts) Clean() {
	*hs = Hosts{}
}

func (hs *Hosts) Add(h Host) {
	(*hs).Hst = append((*hs).Hst, h)
}

// GetAllHosts читает Redis и заполняет срез объектами Host
func (hs *Hosts) GetAllHosts() error {
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = Rdb.Scan(ctx, cursor, "*", poolsize).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			host := NewHost(key)
			if err := host.GetRecordDB(); err != nil {
				return err
			}
			(*hs).Add(host)
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

func NewHost(address string) Host {
	return Host{
		Name:  address,
		IP:    "",
		IsUse: false,
		Stats: "",
		Tab:   "",
	}
}

// InitRedis инициализация Redis
func InitRedis(ctx context.Context) *redis.Client {
	fmt.Println("init DB")
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
