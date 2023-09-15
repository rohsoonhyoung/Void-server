package database

import (
	"sync"
	"time"
)

var (
	HeartBeats     = make(map[string]*HeartBeat)
	HeartBeatMutex sync.RWMutex
)

type HeartBeat struct {
	Ip    string
	Count int64
	Last  time.Time
}

func GetHeartBeats() []*HeartBeat {
	HeartBeatMutex.RLock()
	defer HeartBeatMutex.RUnlock()
	var heartbeats []*HeartBeat
	for _, v := range HeartBeats {
		heartbeats = append(heartbeats, v)
	}
	return heartbeats
}
func GetHeartBeatsByIp(ip string) *HeartBeat {
	HeartBeatMutex.RLock()
	defer HeartBeatMutex.RUnlock()
	ratecounter, ok := HeartBeats[ip]
	if !ok {
		return nil
	}
	return ratecounter
}
func SetHeartBeats(heartbeat *HeartBeat) {
	HeartBeatMutex.Lock()
	defer HeartBeatMutex.Unlock()
	HeartBeats[heartbeat.Ip] = heartbeat
}
