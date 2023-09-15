package database

import (
	"fmt"
	"time"

	"github.com/bujor2711/Void-server/utils"

	null "gopkg.in/guregu/null.v3"
)

var (
	START_WAR       = utils.Packet{0xaa, 0x55, 0x23, 0x00, 0x65, 0x01, 0x00, 0x00, 0x17, 0x00, 0x00, 0x00, 0x10, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0d, 0x00, 0x00, 0x00, 0x10, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xb0, 0x04, 0x00, 0x00, 0x55, 0xaa}
	OrderCharacters []*Character
	ShaoCharacters  []*Character

	WarRequirePlayers  = 10
	OrderPoints        = 10000
	ShaoPoints         = 10000
	CanJoinWar         = false
	WarStarted         = false
	WarStonesPseudoIDs = []uint16{}
	WarStonesIDs       = []int{424203, 424204, 424205, 424206, 424207}
	WarStones          = make(map[int]*WarStone)
	Level              int
)

type WarStone struct {
	PseudoID      uint16
	AIid          int
	NpcID         int
	ConqueredID   int
	ConquereValue int
	NearbyZuhangs []*Character
	NearbyShaos   []*Character
}

func StartWarTimer(prepareWarStart int, level int) {
	Level = level
	min, sec := secondsToMinutes(prepareWarStart)
	msg := fmt.Sprintf("%d minutes %d second after the Great War will start.", min, sec)
	msg2 := fmt.Sprintf("Please participate war by Battle Guard")
	makeAnnouncement(msg)
	makeAnnouncement(msg2)
	if prepareWarStart > 0 {
		time.AfterFunc(time.Second*10, func() {
			StartWarTimer(prepareWarStart-10, level)
		})
	} else {
		StartWar()
	}
}
func secondsToMinutes(inSeconds int) (int, int) {
	minutes := inSeconds / 60
	seconds := inSeconds % 60
	return minutes, seconds
}

type countdown struct {
	t int
	d int
	h int
	m int
	s int
}

func StartWar() {
	OrderPoints = 10000
	ShaoPoints = 10000
	for _, stones := range WarStones {
		stones.ConquereValue = 100
		stones.ConqueredID = 100
	}

	resp := START_WAR
	byteOrders := utils.IntToBytes(uint64(len(OrderCharacters)), 4, false)
	byteShaos := utils.IntToBytes(uint64(len(ShaoCharacters)), 4, false)
	resp.Overwrite(byteOrders, 8)
	resp.Overwrite(byteShaos, 22)
	for _, char := range OrderCharacters {
		char.Socket.Write(resp)
	}
	for _, char := range ShaoCharacters {
		char.Socket.Write(resp)
	}

	CanJoinWar = false
	WarStarted = true
	StartInWarTimer()
}

func (stone *WarStone) RemoveZuhang(c *Character) {
	var arr []*Character
	for _, zhuang := range stone.NearbyZuhangs {
		if zhuang.ID != c.ID {
			arr = append(arr, zhuang)
		}
	}
	stone.NearbyZuhangs = arr
}

func (stone *WarStone) RemoveShao(c *Character) {
	var arr []*Character
	for _, zhuang := range stone.NearbyShaos {
		if zhuang.ID != c.ID {
			arr = append(arr, zhuang)
		}
	}
	stone.NearbyShaos = arr
}
func AddPlayerToGreatWar(c *Character) {
	if !CanJoinWar {
		return
	}
	if (c.Level < 40 || c.Level > 100) && Level == 40 {
		return
	} else if (c.Level < 101 || c.Level > 201) && Level == 101 {
		return
	}

	/*for _, player := range OrderCharacters {
		user, err := FindUserByID(player.UserID)
		if err != nil {
			continue
		}
		user2, err := FindUserByID(c.UserID)
		if err != nil {
			return
		}
		ip1 := strings.Split(user.ConnectedIP, ":")
		ip1x := ip1[0]
		ip2 := strings.Split(user2.ConnectedIP, ":")
		ip2x := ip2[0]

		if ip1x == ip2x {
			c.Socket.Write(messaging.InfoMessage(fmt.Sprintf("You cannot enter with more than one character!")))
			return
		}
	}

	for _, player := range ShaoCharacters {

		user, err := FindUserByID(player.UserID)
		if err != nil {
			continue
		}
		user2, err := FindUserByID(c.UserID)
		if err != nil {
			return
		}
		ip1 := strings.Split(user.ConnectedIP, ":")
		ip1x := ip1[0]
		ip2 := strings.Split(user2.ConnectedIP, ":")
		ip2x := ip2[0]

		if ip1x == ip2x {
			c.Socket.Write(messaging.InfoMessage(fmt.Sprintf("You cannot enter with more than one character!")))
			return
		}
	}*/

	if c.Faction == 1 {
		x := 75.0
		y := 45.0
		c.IsinWar = true
		OrderCharacters = append(OrderCharacters, c)
		data, _ := c.ChangeMap(230, ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
		c.WarKillCount = 0
		c.WarContribution = 0
		c.Socket.Write(data)
	} else {
		x := 81.0
		y := 475.0
		c.IsinWar = true
		ShaoCharacters = append(ShaoCharacters, c)
		data, _ := c.ChangeMap(230, ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
		c.WarKillCount = 0
		c.WarContribution = 0
		c.Socket.Write(data)
	}

}
func GreatWarSchedule() {
	if !WarStarted && !CanJoinWar {
		expiration := null.NewTime(time.Now().Add(time.Second*time.Duration(1)), true)
		hour, minutes := getHour(expiration)
		_, _, day := time.Now().Date()
		if day%2 == 0 { //e zi para
			if hour == 10 && minutes == 0 {
				CanJoinWar = true
				StartWarTimer(int(600), 40)

			}
			if hour == 21 && minutes == 0 {
				CanJoinWar = true
				StartWarTimer(int(600), 40)

			}
		}
		if day%2 != 0 { //e zi impara
			if hour == 11 && minutes == 0 {
				CanJoinWar = true
				StartWarTimer(int(600), 101)
			}
			if hour == 23 && minutes == 0 {
				CanJoinWar = true
				StartWarTimer(int(600), 101)

			}
		}
	}
	time.AfterFunc(time.Second*30, func() {
		GreatWarSchedule()
	})
}
