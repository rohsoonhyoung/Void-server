package dungeon

import (
	"time"

	"github.com/bujor2711/Void-server/database"
	"github.com/bujor2711/Void-server/messaging"
	"github.com/bujor2711/Void-server/utils"
)

var (
	DUNGEON_TIMER = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0xC0, 0x17, 0x09, 0x07, 0x00, 0x00, 0x55, 0xAA}
	//6th = 0x19 -> Monster count x/..

	SeasonDungeonCharacters = make(map[int]*database.Character)
	//DungeonPointsReward     = 1
	IsSeasonDungeon1Closed bool
	IsSeasonDungeon2Closed bool
	IsSeasonDungeon3Closed bool
	IsSeasonDungeon4Closed bool

	TIME_LIMIT = 3600
)

func StartSeasonDungeon(s *database.Socket) {

	server := 1
	if !IsSeasonDungeon1Closed {
		IsSeasonDungeon1Closed = true
	} else if !IsSeasonDungeon2Closed {
		server = 2
		IsSeasonDungeon2Closed = true
	} else if !IsSeasonDungeon3Closed {
		server = 3
		IsSeasonDungeon3Closed = true
	} else if !IsSeasonDungeon4Closed {
		server = 4
		IsSeasonDungeon4Closed = true
	} else {
		msg := messaging.InfoMessage("All dungeons are full at this moment, come back later. ")
		s.Write(msg)
		return
	}
	s.Character.IsDungeon = true
	s.Character.Socket.User.ConnectedServer = server
	s.User.SelectedServerID = server
	data, _ := s.Character.ChangeMap(212, nil)
	s.Write(data)

	go StartSeasonTimer(s, 212, TIME_LIMIT)
	go database.CountSeasonCave(server)
	go SetSeasonDungeonOpenAfterTime(server)

}
func SetSeasonDungeonOpenAfterTime(server int) {
	time.Sleep(time.Second * time.Duration(TIME_LIMIT))
	if server == 1 {
		IsSeasonDungeon1Closed = false
	} else if server == 2 {
		IsSeasonDungeon2Closed = false
	} else if server == 3 {
		IsSeasonDungeon3Closed = false
	} else if server == 4 {
		IsSeasonDungeon4Closed = false
	}
}

func StartSeasonTimer(s *database.Socket, mapID int16, seconds int) {
	resp := DUNGEON_TIMER
	resp.Overwrite(utils.IntToBytes(uint64(seconds), 4, true), 6)
	s.Write(DUNGEON_TIMER)

	time.AfterFunc(time.Second*time.Duration(seconds), func() {
		if s.Character.Map == 212 && s.Character.IsOnline {
			resp := utils.Packet{}
			resp.Concat(messaging.InfoMessage("Your time has ended. Come again when you are stronger. Teleporting to safe zone."))
			data, _ := s.Character.ChangeMap(1, nil)
			resp.Concat(data)
			s.Write(resp)
			s.Character.IsDungeon = false
		}
	})

}
