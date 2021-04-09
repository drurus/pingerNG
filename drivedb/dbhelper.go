package drivedb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ctx = context.Background()
	Rdb *redis.Client
)

const (
	poolsize    = 100        // размер пула для чтения из БД
	statsize    = 75         // кол-во сохраняемых значений статистики
	Sk_template = "loss,rtt" // список стат-ключей
)

func init() {
	Rdb = InitRedis(ctx)
}

type Host struct {
	Name  string `json:"name"`  // hostname
	IP    string `json:"ip"`    // адрес
	IsUse bool   `json:"isuse"` // флаг обработки, 0 - false
	Stats []Stat `json:"stats"` // результат обработки
	// Stats string `json:"stats"` // результат обработки
	Tab string `json:"tab"` // имя таба
}

type Stat struct {
	Type   string    `json:"type"`
	Values []float64 `json:"values"`
	// Values []string `json:"values"`
}

type Hosts struct {
	Hst []Host
}

// type Hosts struct {
// 	Data []Host `json:"data"`
// 	Err  string `json:"err"`
// }

func statsToString(s *[]Stat) string {
	// Marshal
	st := *s
	var sout []byte
	for _, stat := range st {
		js, err := json.Marshal(stat)
		if err != nil {
			fmt.Printf("\terror convert Stat to String: %s\n", err)
			return ""
		}
		sout = append(sout, js...)
	}
	return string(sout)
}

// ToStrings выводит структуру Host как []string
func (h *Host) ToStrings() []string {
	sc := make([]string, 0)
	sc = append(sc, "Name", h.Name)
	sc = append(sc, "IP", h.IP)
	sc = append(sc, "IsUse", strconv.FormatBool((h.IsUse)))
	sc = append(sc, "Stats", statsToString(&h.Stats))
	// sc = append(sc, "Stats", h.Stats)
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
	//-	(*h).Stats = hh["Stats"]
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

// Rsk возвращает список по стат-ключу
func (h *Host) Rsk(key string) []float64 {
	// func (h *Host) Rsk(key string) []string {
	sl := Rdb.LRange(ctx, key, 0, statsize-1)
	if sl.Err() != nil {
		return []float64{}
	}
	if sval, err := sl.Result(); err != nil {
		return []float64{}
	} else {
		// return sval
		var sfval []float64
		for _, val := range sval {
			if conv, err := strconv.ParseFloat(val, 64); err != nil {
				sfval = append(sfval, 0.0)
			} else {
				sfval = append(sfval, conv)
			}
		}
		return sfval
	}
}

// psk записывает список по стат-ключу
func (h *Host) psk(key string, vals []string) error {
	r := Rdb.LPush(ctx, key, vals)
	if r.Err() != nil {
		return r.Err()
	}
	if _, err := r.Result(); err != nil {
		fmt.Printf("error PUSH %s %q: %s", key, vals, err)
		return err
	} else {
		return nil
	}
}

// tsk обрезает список по стат-ключу
func (h *Host) tsk(key string) error {
	// sl := Rdb.LTrim(ctx, key, 1, statsize)
	sl := Rdb.LTrim(ctx, key, 0, statsize-1)
	if sl.Err() != nil {
		return sl.Err()
	}
	if _, err := sl.Result(); err != nil {
		return err
	} else {
		return nil
	}
}

// SaveStats сохраняет списки статистики
//  Host.Stats должна содержать статистические ключи
func (h *Host) SaveStats(statkey string, sval []string) error {
	// -1) получить и распарсить доступные стат-ключи
	// 2) загрузить из БД значения стат-ключей RPUSH, LTRIM, LRANGE
	// 3) сохранить/добавить новые значения
	// fmt.Printf("\n------SaveStats for %s: statkey: %s, sval: %v\n", *&h.Name, statkey, sval)
	skeys := strings.Split(Sk_template, ",")
	// skeys := strings.Split((*h).Stats, ",")
	flagapply := false // доступность ключа
	for _, key := range skeys {
		if statkey == key {
			flagapply = true
		}
	}
	if !flagapply {
		return fmt.Errorf("\tkey %q is not avaliable\n", statkey)
	}
	// if len(skeys) <= 0 {
	// 	return nil
	// }

	realskey := (*h).Name + "_" + statkey
	// fmt.Println("realskey: ", realskey)
	// загрузить // LRANGE 4.2.2.1_rtt 0 statsize-1
	kk := (*h).Rsk(realskey)
	// fmt.Println("-->> ", kk)
	var skk []string
	for _, val := range kk {
		skk = append(skk, strconv.FormatFloat(val, 'f', 2, 64))
	}
	// fmt.Printf("SKK: %s; SVAL: %s\n", skk, sval)

	// TODO skk = append(sval, skk...)

	// fmt.Println("SKK append: ", skk)
	// kk = append(kk, sval...)
	// добавить // RPUSH 4.2.2.1_rtt "12.68"
	if err := (*h).psk(realskey, sval); err != nil {
		// if err := (*h).psk(realskey, skk); err != nil {
		// if err := (*h).psk(realskey, kk); err != nil {
		return err
	}
	// обрезать // LTRIM 4.2.2.1_rtt 0 statsize-1
	if err := (*h).tsk(realskey); err != nil {
		return err
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
	// обработка паники (например не получен IP адрес)
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("ошибка обработки хоста: %s\n", err)
		}
	}()

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = Rdb.Scan(ctx, cursor, "*", poolsize).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {

			// если ключ статистики - пропустить
			ok := checkKeyAsStats(key)
			if ok {
				// fmt.Printf("\tStatistics key: %s\n", statkey)
				continue
			}

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

// checkKeyAsStats проверяет ключ на соответствие ключу статистики
func checkKeyAsStats(s string) bool {
	spl := strings.Split(s, "_")
	if len(spl) == 2 {
		return true
	}
	return false
}

func NewHost(address string) Host {
	return Host{
		Name:  address,
		IP:    "",
		IsUse: false,
		// Stats: "loss,rtt", // ключи статистики по умолчанию
		Stats: []Stat{}, // ключи статистики по умолчанию
		Tab:   "",
	}
}

// InitRedis инициализация Redis
func InitRedis(ctx context.Context) *redis.Client {
	fmt.Print("init DB ...")
	r := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	for {
		_, err := r.Ping(ctx).Result()
		if err == nil {
			fmt.Println("OK")
			return r
		}
		<-time.After(time.Millisecond * 750)
	}
}
