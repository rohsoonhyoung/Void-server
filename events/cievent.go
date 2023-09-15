package events

import (
	"fmt"
	"time"

	"github.com/bujor2711/Void-server/database"
	"github.com/bujor2711/Void-server/nats"
	"github.com/bujor2711/Void-server/utils"

	null "gopkg.in/guregu/null.v3"
)

var (
	ANNOUNCEMENT = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x71, 0x06, 0x00, 0x55, 0xAA}
	mob          *database.AI
	respawns     = 4
	isCiRunning  = false
)

func CiEventSchedule() {

	if !isCiRunning {
		hour, minutes := getHour(null.NewTime(time.Now(), true))
		if (hour == 4 && minutes == 0) || (hour == 12 && minutes == 0) || (hour == 20 && minutes == 0) {
			StartCiEventCountdown(600)
			isCiRunning = true

		}

	}
	time.AfterFunc(time.Minute, func() {
		CiEventSchedule()
	})
}
func StartCiEventCountdown(cd int) {
	if cd >= 120 {
		msg := fmt.Sprintf("Central Island event will start in %d minutes.", cd/60)
		MakeAnnouncement(msg)
		time.AfterFunc(time.Second*60, func() {
			StartCiEventCountdown(cd - 60)
		})
	} else if cd > 0 {
		msg := fmt.Sprintf("Central Island event will start in %d seconds.", cd)
		MakeAnnouncement(msg)
		time.AfterFunc(time.Second*10, func() {
			StartCiEventCountdown(cd - 10)
		})
	}
	if cd <= 0 {
		StartCiEvent()
	}
}

func StartCiEvent() {
	if mob == nil {
		spawnMob()
	}
	if respawns <= 0 {
		msg := "Central Island event finished, thank you for participation"
		MakeAnnouncement(msg)
		respawns = 4
		isCiRunning = false
		return
	} else {
		time.AfterFunc(time.Second*5, func() {
			if mob.IsDead {
				spawnMob()
			}
			StartCiEvent()
		})
	}

}
func spawnMob() {
	respawns--
	id := 77

	npcPos := database.GetNPCPosByID(id)
	if npcPos == nil {
		return
	}
	npc, ok := database.GetNpcInfo(npcPos.NPCID)
	if !ok || npc == nil {
		return
	}

	msg := fmt.Sprintf("%s is roaring.", npc.Name)
	MakeAnnouncement(msg)

	ai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: 10, PosID: npcPos.ID, RunningSpeed: 10, Server: 1, WalkingSpeed: 5, Once: true, CanAttack: true, Faction: 0, IsDead: false}
	ai.OnSightPlayers = make(map[int]interface{})

	points := []string{"105,401", "259,397", "383,369", "409,239", "339,159", "221,137", "129,119", "99,191", "117,273", "189,339", "293,343", "329,271", "273,195", "181,189", "203,345"}
	min := utils.RandInt(0, int64(len(points)-1))
	max := utils.RandInt(0, int64(len(points)-1))
	for min == max {
		min = utils.RandInt(0, int64(len(points)-1))
		max = utils.RandInt(0, int64(len(points)-1))
	}

	npcPos.MinLocation = points[min]
	npcPos.MaxLocation = points[max]

	minLoc := database.ConvertPointToLocation(npcPos.MinLocation)
	maxLoc := database.ConvertPointToLocation(npcPos.MaxLocation)
	loc := utils.Location{X: utils.RandFloat(minLoc.X, maxLoc.X), Y: utils.RandFloat(minLoc.Y, maxLoc.Y)}
	ai.NPCpos = npcPos
	ai.Coordinate = loc.String()
	ai.TargetLocation = *database.ConvertPointToLocation(ai.Coordinate)
	database.GenerateIDForAI(ai)
	ai.OnSightPlayers = make(map[int]interface{})
	ai.Handler = ai.AIHandler

	database.AIsByMap[ai.Server][ai.Map] = append(database.AIsByMap[ai.Server][ai.Map], ai)
	database.AIs[ai.ID] = ai
	mob = ai
	go ai.Handler()

}

func MakeAnnouncement(msg string) {
	length := int16(len(msg) + 3)

	resp := ANNOUNCEMENT
	resp.SetLength(length)
	resp[6] = byte(len(msg))
	resp.Insert([]byte(msg), 7)

	p := nats.CastPacket{CastNear: false, Data: resp}
	p.Cast()
}
func getHour(date null.Time) (int, int) {
	if date.Valid {
		hours, minutes, _ := date.Time.Clock()
		return hours, minutes
	}
	return 0, 0
}
