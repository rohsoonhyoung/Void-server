package dungeon

import (
	"time"

	"github.com/bujor2711/Void-server/database"
	"github.com/bujor2711/Void-server/messaging"
	"github.com/bujor2711/Void-server/utils"
)

var (
	IsDungeonClosed = false
	YY_TIME_LIMIT   = 30
)

func StartYingYang(party *database.Party) {

	server := 1
	if !IsDungeonClosed {
		IsDungeonClosed = true
	} else {
		msg := messaging.InfoMessage("All dungeons are full at this moment, come back later.")
		party.Leader.Socket.Write(msg)
		return
	}
	party.Leader.Socket.User.ConnectedServer = server
	party.Leader.Socket.User.SelectedServerID = server
	data, _ := party.Leader.ChangeMap(243, nil)
	party.Leader.Socket.Write(data)
	for _, member := range party.Members {

		member.Character.Socket.User.ConnectedServer = server
		member.Character.Socket.User.SelectedServerID = server
		member.IsDungeon = true
		data, _ := member.Character.ChangeMap(243, nil)
		member.Character.Socket.Write(data)
		go StartTimerYingYang(member.Character.Socket, 900)
	}

	go StartTimerYingYang(party.Leader.Socket, YY_TIME_LIMIT)
	go SetDungeonOpenAfterTime(server)
	go database.CountYingYangMobs(party.Leader.Map)

}
func SetDungeonOpenAfterTime(server int) {
	time.Sleep(time.Minute * time.Duration(YY_TIME_LIMIT))

	IsDungeonClosed = false
}
func StartTimerYingYang(s *database.Socket, minutes int) {
	resp := DUNGEON_TIMER
	resp.Overwrite(utils.IntToBytes(uint64(minutes*60), 4, true), 6)
	s.Write(DUNGEON_TIMER)

	time.AfterFunc(time.Minute*time.Duration(YY_TIME_LIMIT), func() {
		if s.Character.Map == 243 && s.Character.IsOnline {
			resp := utils.Packet{}
			resp.Concat(messaging.InfoMessage("Your time has ended. Come again when you are stronger. Teleporting to safe zone."))
			coordinate := &utils.Location{X: 37, Y: 453}
			data, _ := s.Character.ChangeMap(17, coordinate)
			resp.Concat(data)
			s.Write(resp)
			s.Character.IsDungeon = false
		}
	})
}
