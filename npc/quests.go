package npc

import (
	"log"

	"github.com/bujor2711/Void-server/database"
	"github.com/bujor2711/Void-server/utils"
)

func GetQuestNPCMenu(npcID, textID, index int, actions []int, char *database.Character) []byte {
	resp := NPC_MENU
	resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 6)         // npc id
	resp.Insert(utils.IntToBytes(uint64(textID), 4, true), 10)       // text id
	resp.Insert(utils.IntToBytes(uint64(len(actions)), 1, true), 14) // action length

	counter, length := 15, int16(11)
	for i, action := range actions {
		resp.Insert(utils.IntToBytes(uint64(action), 4, true), counter) // action
		counter += 4
		char.QuestActions = append(char.QuestActions, action)
		resp.Insert(utils.IntToBytes(uint64(npcID), 2, true), counter) // npc id
		counter += 2

		actIndex := int(index) + (i+1)<<(len(actions)*3)
		resp.Insert(utils.IntToBytes(uint64(actIndex), 2, true), counter) // action index
		counter += 2

		length += 8
	}

	resp.SetLength(length)
	return resp
}
func getQuestMenuIds(npcID int, c *database.Character) []int {
	var questMenus []int
	AllNPCQuest, _ := database.FindQuestByNpcID(npcID)
	for _, q := range AllNPCQuest {
		Playerq, err := database.FindPlayerQuestByID(q.ID, c.ID)
		if err != nil {
			log.Print("ErrorWithOpen")
		}
		if Playerq != nil && Playerq.QuestState == 3 && q.NPCID == int64(npcID) {
			questMenus = append(questMenus, q.MenuID)
		} else if Playerq != nil && Playerq.QuestState == 4 && q.FinishNPC == npcID {
			questMenus = append(questMenus, q.MenuID)
		}
	}
	return questMenus
}

func finishQuest(questID int, char *database.Character) ([]byte, error) {
	hasAllItem := true
	mQuest := database.QuestsList[questID]
	questReqItems, _ := mQuest.GetQuestReqItems()
	questRewards, _ := mQuest.GetQuestRewItems()
	slotCount, err := char.FindFreeSlots(len(questRewards))
	if err != nil {
		return nil, err
	}
	fullslot := len(questRewards) - len(questReqItems)
	if len(slotCount) >= fullslot {
		for _, items := range questReqItems {
			slotID, _, _ := char.FindItemInInventory(nil, items.ItemID)
			slots, err := char.InventorySlots()
			if err != nil {
				return nil, err
			}
			item := slots[slotID]
			if item.Quantity < uint(items.ItemCount) {
				hasAllItem = false
			}
		}
		if hasAllItem {
			for _, items := range questReqItems {
				slotID, _, _ := char.FindItemInInventory(nil, items.ItemID)
				itemData := char.DecrementItem(slotID, uint(items.ItemCount))
				char.Socket.Write(*itemData)
			}
			resp := utils.Packet{}
			data, levelUp := char.AddExp(int64(mQuest.RewardExp))
			char.LootGold(uint64(mQuest.RewardGold))
			resp.Concat(char.GetGold())
			if len(questRewards) > 0 {
				for _, item := range questRewards {
					itemData, _, err := char.AddItem(&database.InventorySlot{ItemID: item.ItemID, Quantity: uint(item.ItemCount)}, -1, false)
					if err != nil {
						return nil, err
					} else if resp == nil {
						return nil, nil
					}
					resp.Concat(*itemData)
				}
			}
			if levelUp {
				statData, err := char.GetStats()
				if err == nil && char.Socket != nil {
					char.Socket.Write(statData)
				}
			}

			if char.Socket != nil {
				char.Socket.Write(data)
			}

			playerquest, err := database.FindPlayerQuestByID(mQuest.ID, char.ID)
			if err != nil {
				return nil, err
			}
			char.LoadQuests(mQuest.ID, 2)
			playerquest.QuestState = 2
			playerquest.Update()
			return resp, nil
		}
	} else {
		log.Print("There is not enough space for the player.")
	}
	return nil, nil
}
