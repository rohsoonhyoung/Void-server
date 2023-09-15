package database

import (
	"fmt"
	"log"
	"time"

	"github.com/bujor2711/Void-server/messaging"
	"github.com/bujor2711/Void-server/utils"
)

var (
	TIMER_MENU     = utils.Packet{0xAA, 0x55, 0x08, 0x00, 0x65, 0x03, 0x00, 0x55, 0xAA}
	WAR_SCOREPANEL = utils.Packet{0xAA, 0x55, 0x30, 0x00, 0x65, 0x06, 0x55, 0xAA}
)

func StartInWarTimer() {
	timein := time.Now().Add(time.Minute * 20)
	deadtime := timein.Format(time.RFC3339)

	v, err := time.Parse(time.RFC3339, deadtime)
	if err != nil {
		log.Print(err)
		//os.Exit(1)
	}

	for range time.Tick(1 * time.Second) {
		checkMembersInGreatWar()
		timeRemaining := getTimeRemaining(v)
		if timeRemaining.t <= 0 || OrderPoints <= 0 || ShaoPoints <= 0 || len(OrderCharacters) == 0 || len(ShaoCharacters) == 0 {
			finishGreatWar()
			break
		}
		data := utils.IntToBytes(uint64(timeRemaining.t), 4, true)
		shaoStones := 0
		ZuhangStones := 0
		index := 7
		resp := TIMER_MENU
		byteOrders := utils.IntToBytes(uint64(len(OrderCharacters)), 4, true)
		ordersPoint := utils.IntToBytes(uint64(OrderPoints), 4, true)
		byteShaos := utils.IntToBytes(uint64(len(ShaoCharacters)), 4, true)
		shaoPoint := utils.IntToBytes(uint64(ShaoPoints), 4, true)
		for _, stones := range WarStones {
			if stones.ConqueredID == 1 {
				ShaoPoints -= 2
				ZuhangStones++
			} else if stones.ConqueredID == 2 {
				OrderPoints -= 2
				shaoStones++
			}
		}
		resp.Insert(byteOrders, index)
		index += 4
		resp.Insert(ordersPoint, index)
		index += 4
		resp.Insert([]byte{0x00, 0x00, 0x00, 0x00}, index)
		index += 4
		if ZuhangStones > 0 {
			resp.Insert(utils.IntToBytes(uint64(ZuhangStones), 1, false), index)
			index++
			for _, stones := range WarStones {
				if stones.ConqueredID == 1 {
					resp.Insert(utils.IntToBytes(uint64(stones.NpcID), 4, true), index)
					index += 4
				}
			}
			resp.Insert([]byte{0x00}, index)
			index++
		} else {
			resp.Insert([]byte{0x00, 0x00}, index) //IDE JÖN MAJD HOGY KINEK HÁNY KÖVE VAN
			index += 2
		}
		resp.Insert(byteShaos, index)
		index += 4
		resp.Insert(shaoPoint, index)
		index += 4
		resp.Insert([]byte{0x00, 0x00, 0x00, 0x00}, index)
		index += 4
		if shaoStones >= 1 {
			resp.Insert(utils.IntToBytes(uint64(shaoStones), 1, false), index)
			index++
			for _, stones := range WarStones {
				if stones.ConqueredID == 2 {
					resp.Insert(utils.IntToBytes(uint64(stones.NpcID), 4, true), index)
					index += 4
				}
			}
		} else {
			resp.Insert([]byte{0x00}, index-2)
			index++
		}
		resp.Insert(data, index)
		index += 4
		/*resp.Insert(data2, index)
		index++*/
		length := index - 4
		resp.SetLength(int16(length))
		for _, char := range OrderCharacters {
			if char.IsOnline {
				char.Socket.Write(resp)
			}
		}
		for _, char := range ShaoCharacters {
			if char.IsOnline {
				char.Socket.Write(resp)
			}
		}
		for _, stones := range WarStones {
			if len(stones.NearbyZuhangs) > len(stones.NearbyShaos) {
				if stones.ConquereValue > 0 {
					stones.ConquereValue--
				}
				if stones.ConquereValue >= 0 && stones.ConquereValue <= 30 {
					stones.ConqueredID = 1
				} else if stones.ConquereValue > 170 {
					stones.ConqueredID = 0
				}
			} else if len(stones.NearbyShaos) > len(stones.NearbyZuhangs) {
				if stones.ConquereValue < 200 {
					stones.ConquereValue++
				}
				if stones.ConquereValue >= 170 && stones.ConquereValue <= 200 {
					stones.ConqueredID = 2
				} else if stones.ConquereValue < 30 {
					stones.ConqueredID = 0
				}
			}
		}
	}
}

func finishGreatWar() {
	checkMembersInGreatWar()
	greatWarRewards()
	resp := WAR_SCOREPANEL
	index := 6
	if OrderPoints < ShaoPoints {
		resp.Insert([]byte{0x00, 0x28, 0x00}, index)
	} else {
		resp.Insert([]byte{0x01, 0x28, 0x00}, index)
	}
	index += 3
	for _, char := range OrderCharacters {
		resp.Insert(utils.IntToBytes(uint64(len(char.Name)), 1, false), index)
		index++
		resp.Insert([]byte(char.Name), index)
		index += len(char.Name)
		resp.Insert(utils.IntToBytes(uint64(char.Faction), 1, false), index)
		index++
		data := utils.IntToBytes(uint64(char.WarContribution), 2, true)
		resp.Insert(data, index)
		index += 2
		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
		data2 := utils.IntToBytes(uint64(char.WarKillCount), 2, true)
		resp.Insert(data2, index)
		index += 2
		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
	}
	for _, char := range ShaoCharacters {
		resp.Insert(utils.IntToBytes(uint64(len(char.Name)), 1, false), index)
		index++
		resp.Insert([]byte(char.Name), index)
		index += len(char.Name)
		resp.Insert(utils.IntToBytes(uint64(char.Faction), 1, false), index)
		index++
		data := utils.IntToBytes(uint64(char.WarContribution), 2, true)
		resp.Insert(data, index)
		index += 2
		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
		data2 := utils.IntToBytes(uint64(char.WarKillCount), 2, true)
		resp.Insert(data2, index)
		index += 2
		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
	}
	length := index - 4
	resp.SetLength(int16(length))
}

func getTimeRemaining(t time.Time) countdown {
	currentTime := time.Now()
	difference := t.Sub(currentTime)

	total := int(difference.Seconds())
	days := int(total / (60 * 60 * 24))
	hours := int(total / (60 * 60) % 24)
	minutes := int(total/60) % 60
	seconds := int(total % 60)
	return countdown{
		t: total,
		d: days,
		h: hours,
		m: minutes,
		s: seconds,
	}
}
func greatWarRewards() {
	if OrderPoints > ShaoPoints { //zhuang won
		for _, c := range OrderCharacters { //give item to all zhuangs
			if c == nil {
				continue
			}
			item := &InventorySlot{ItemID: 200001116, Quantity: uint(1)}
			r, _, err := c.AddItem(item, -1, false)
			if err == nil {
				c.Socket.Write(*r)
			}
			item = &InventorySlot{ItemID: 100080299, Quantity: uint(1)}
			r, _, err = c.AddItem(item, -1, false)
			if err == nil {
				c.Socket.Write(*r)
			}
			c.Socket.Stats.Honor += 50
			c.Socket.Stats.Update()
			c.Socket.Write(messaging.InfoMessage(fmt.Sprintf("You acquired 50 Honor points.")))
			stat, _ := c.GetStats()
			c.Socket.Write(stat)

			go time.AfterFunc(time.Second*10, func() {
				gomap, _ := c.ChangeMap(1, nil)
				c.Socket.Write(gomap)
				c.IsinWar = false
			})
		}
		for _, c := range ShaoCharacters { //give item to all shaos
			if c == nil {
				continue
			}
			item := &InventorySlot{ItemID: 200001115, Quantity: uint(1)}
			r, _, err := c.AddItem(item, -1, false)
			if err == nil {
				c.Socket.Write(*r)
			}
			item = &InventorySlot{ItemID: 100080300, Quantity: uint(1)}
			r, _, err = c.AddItem(item, -1, false)
			if err == nil {
				c.Socket.Write(*r)
			}
			go time.AfterFunc(time.Second*10, func() {
				gomap, _ := c.ChangeMap(1, nil)
				c.Socket.Write(gomap)
				c.IsinWar = false
			})
		}
	} else { // shao won
		for _, c := range OrderCharacters { //give item to all zhuangs
			if c == nil {
				continue
			}
			item := &InventorySlot{ItemID: 200001115, Quantity: uint(1)}
			r, _, err := c.AddItem(item, -1, false)
			if err == nil {
				c.Socket.Write(*r)
			}
			item = &InventorySlot{ItemID: 100080300, Quantity: uint(1)}
			r, _, err = c.AddItem(item, -1, false)
			if err == nil {
				c.Socket.Write(*r)
			}
			go time.AfterFunc(time.Second*10, func() {
				gomap, _ := c.ChangeMap(1, nil)
				c.Socket.Write(gomap)
				c.IsinWar = false
			})
		}
		for _, c := range ShaoCharacters { //give item to all shaos
			if c == nil {
				continue
			}
			item := &InventorySlot{ItemID: 200001116, Quantity: uint(1)}
			r, _, err := c.AddItem(item, -1, false)
			if err == nil {
				c.Socket.Write(*r)
			}
			item = &InventorySlot{ItemID: 100080299, Quantity: uint(1)}
			r, _, err = c.AddItem(item, -1, false)
			if err == nil {
				c.Socket.Write(*r)
			}
			c.Socket.Stats.Honor += 50
			c.Socket.Stats.Update()
			c.Socket.Write(messaging.InfoMessage(fmt.Sprintf("You acquired 50 Honor points.")))
			stat, _ := c.GetStats()
			c.Socket.Write(stat)
			go time.AfterFunc(time.Second*10, func() {
				gomap, _ := c.ChangeMap(1, nil)
				c.Socket.Write(gomap)
				c.IsinWar = false
			})
		}
	}
	WarStarted = false
}
func checkMembersInGreatWar() {
	OrderCharacters = nil
	ShaoCharacters = nil
	for _, member := range FindCharactersInMap(230) {
		if member.Faction == 1 {
			OrderCharacters = append(OrderCharacters, member)
		}
		if member.Faction == 2 {
			ShaoCharacters = append(ShaoCharacters, member)
		}
	}
}
