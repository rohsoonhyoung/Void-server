package dungeon

import (
	"fmt"
	"time"

	"github.com/bujor2711/Void-server/database"
	"github.com/bujor2711/Void-server/nats"
	"github.com/bujor2711/Void-server/utils"
)

var (
	ANNOUNCEMENT     = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x71, 0x06, 0x00, 0x55, 0xAA}
	NcashMobs        []int
	NcashMobsPosIDs  = []int{5335, 5336, 5337, 5338}
	SeasonMobsPosIDs = []int{5344, 5345}
)

func SpawnRandomNcashBosses() {
	minutes := utils.RandInt(600, 1440)
	posId := int(utils.RandInt(0, int64(len(NcashMobsPosIDs))))
	posId = NcashMobsPosIDs[posId]

	time.AfterFunc(time.Minute*time.Duration(minutes), func() {
		go SpawnNcashBoss("", posId, true)
		go SpawnRandomNcashBosses()
	})
}

func SpawnSeasonBoss() {
	minutes := utils.RandInt(1350, 1440)
	posId := int(utils.RandInt(0, int64(len(SeasonMobsPosIDs))))
	posId = SeasonMobsPosIDs[posId]
	go SpawnNcashBoss("", posId, true)

	time.AfterFunc(time.Minute*time.Duration(minutes), func() {
		go SpawnNcashBoss("", posId, true)
		go SpawnRandomNcashBosses()
	})
}

func SpawnNcashBoss(coordinate string, posId int, announce bool) {
	npcPos := database.GetNPCPosByID(int(posId))
	if npcPos == nil {
		return
	}
	npc, ok := database.GetNpcInfo(npcPos.NPCID)
	if !ok {
		return
	}

	ai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: npcPos.MapID, PosID: npcPos.ID, RunningSpeed: 10, Server: 1, WalkingSpeed: 5, Once: true}
	database.GenerateIDForAI(ai)
	ai.OnSightPlayers = make(map[int]interface{})

	minLoc := database.ConvertPointToLocation(npcPos.MinLocation)
	maxLoc := database.ConvertPointToLocation(npcPos.MaxLocation)
	loc := utils.Location{X: utils.RandFloat(minLoc.X, maxLoc.X), Y: utils.RandFloat(minLoc.Y, maxLoc.Y)}
	if coordinate != "" {
		ai.Coordinate = coordinate
	}
	ai.Coordinate = loc.String()
	ai.Handler = ai.AIHandler
	go ai.Handler()

	if announce {
		msg := fmt.Sprintf("%s is roaring.", npc.Name)
		makeAnnouncement(msg)

	}

	database.AIsByMap[ai.Server][npcPos.MapID] = append(database.AIsByMap[ai.Server][npcPos.MapID], ai)
	database.AIs[ai.ID] = ai

	NcashMobs = append(NcashMobs, ai.ID)
}
func makeAnnouncement(msg string) {
	length := int16(len(msg) + 3)

	resp := ANNOUNCEMENT
	resp.SetLength(length)
	resp[6] = byte(len(msg))
	resp.Insert([]byte(msg), 7)

	p := nats.CastPacket{CastNear: false, Data: resp}
	p.Cast()
}
