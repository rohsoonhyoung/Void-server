package player

import (
	"github.com/bujor2711/Void-server/database"
	"github.com/bujor2711/Void-server/messaging"
	"github.com/bujor2711/Void-server/nats"
	"github.com/bujor2711/Void-server/utils"
)

type LootHandler struct {
}

func (h *LootHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}

	u := s.User
	if u == nil {
		return nil, nil
	}

	c.Looting.Lock()
	defer c.Looting.Unlock()
	resp := utils.Packet{}

	dropID := uint16(utils.BytesToInt(data[7:9], true))
	drop := database.GetDrop(s.User.ConnectedServer, s.Character.Map, dropID)
	if drop != nil && drop.Item != nil && (drop.Claimer == nil || drop.Claimer.ID == s.Character.ID) {
		if drop.Item.ItemID == 0 {
			return nil, nil
		}
		if drop.Item.ItemID == 99059990 || drop.Item.ItemID == 99059991 || drop.Item.ItemID == 99059992 {
			inv, _ := c.InventorySlots()
			if inv[11].ItemID != 0 {
				return messaging.InfoMessage("Clear first invenotry slot!"), nil
			}
			database.FactionCapturedFlagNotification()
		}

		d, _, err := c.AddItem(drop.Item, -1, true)
		if err != nil {
			return nil, err
		} else if d == nil {
			return nil, nil
		}

		database.RemoveFromDropRegister(s.User.ConnectedServer, s.Character.Map, dropID)
		resp.Concat(*d)
	}

	r := database.DROP_DISAPPEARED
	r.Insert(utils.IntToBytes(uint64(dropID), 2, true), 6) //drop id

	p := nats.CastPacket{CastNear: true, DropID: int(dropID), Data: r, Type: nats.DROP_DISAPPEAR}
	p.Cast()

	resp.Concat(r)
	return resp, nil
}
