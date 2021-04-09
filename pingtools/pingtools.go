package pingtools

import (
	"time"

	"github.com/go-ping/ping"
)

const (
	ping_count = 7
)

// fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
// fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
// 	stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
// fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
// 	stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)

func ProcessPing(addr string) (*ping.Statistics, error) {
	pinger, err := ping.NewPinger(addr)
	// pinger, err := ping.NewPinger((*a).IP)
	if err != nil {
		return &ping.Statistics{
			PacketLoss: 100,
		}, err
		// return &ping.Statistics{}, err
	}
	pinger.Count = ping_count
	pinger.Timeout = time.Millisecond * ping_count * 1500
	pinger.SetPrivileged(true)
	err = pinger.Run()
	if err != nil {
		return pinger.Statistics(), err
	}
	return pinger.Statistics(), nil
}
