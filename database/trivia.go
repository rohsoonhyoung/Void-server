package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/bujor2711/Void-server/messaging"
	"github.com/bujor2711/Void-server/utils"

	"gopkg.in/gorp.v1"
)

var (
	TriviaEventStarted  = false
	TriviaEventCounting = false
	QuestionsList       = make(map[int]*QuestionsItem)
	QuestionAsked       *QuestionsItem
	TriviaParticipants  []*Character
	QuestionRound       = 0
	QuestionMaxRounds   = 5
	EventRound          = 1
	EventMaxRounds      = 3

	NPC_MENU         = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x57, 0x02, 0x55, 0xAA}
	INCORRECT_ANSWER = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x72, 0x11, 0x01, 0x55, 0xAA}
)

type QuestionsItem struct {
	ID       int    `db:"id"`
	Question string `db:"question"`
	Answer   bool   `db:"answer"`
}

func StartTriviaCounter(cd int) {
	TriviaEventCounting = true

	if cd >= 120 {
		checkMembersInFactionWarMap()
		msg := fmt.Sprintf("Trivia event Round %d will start in %dmin. Register to the event at Event Manager Mae.", EventRound, cd/60)
		makeAnnouncement(msg)
		time.AfterFunc(time.Second*60, func() {
			StartTriviaCounter(cd - 60)
		})
	} else if cd > 0 {
		checkMembersInFactionWarMap()
		msg := fmt.Sprintf("Trivia event Round %d will start in %dsec. Register to the event at Event Manager Mae.", EventRound, cd)
		makeAnnouncement(msg)
		time.AfterFunc(time.Second*10, func() {
			StartTriviaCounter(cd - 10)
		})
	}
	if cd <= 0 {
		TriviaEventCounting = false
		TriviaEventStarted = true
		QuestionRound = 0
		nextQuestion()
	}
}

func nextQuestion() {
	QuestionRound++
	rand := int(utils.RandInt(1, int64(len(QuestionsList))))
	QuestionAsked = QuestionsList[rand]
	for _, member := range TriviaParticipants {
		c := member
		c.TriviaAnswer = 3
		c.removeTriviaItems()
		itemData, _, err := c.AddItem(&InventorySlot{ItemID: 17502588, Quantity: 1}, -1, false) //true
		if err != nil {
			continue
		}
		c.Socket.Write(*itemData)
		itemData, _, err = c.AddItem(&InventorySlot{ItemID: 17502584, Quantity: 1}, -1, false) //false
		if err != nil {
			continue
		}
		c.Socket.Write(*itemData)

		c.Socket.Write(messaging.InfoMessage(QuestionAsked.Question))
	}

	time.AfterFunc(time.Second*20, func() {
		endRound()
	})
}
func endRound() {
	for _, member := range TriviaParticipants {
		c := member
		c.removeTriviaItems()

		if (member.TriviaAnswer == 1 && QuestionAsked.Answer) || (member.TriviaAnswer == 0 && !QuestionAsked.Answer) {
			c.Socket.Write(messaging.InfoMessage("Your answer was correct."))
			if QuestionRound <= QuestionMaxRounds {
				c.Socket.Write(messaging.InfoMessage("Next question will be displayed in 5 seconds..."))
			}
		} else if member.TriviaAnswer == 3 {
			c.Socket.Write(messaging.InfoMessage("Sorry, your are too slow to answer. Answer faster next time."))
			go c.RemoveFromTrivia()

		} else {
			c.Socket.Write(INCORRECT_ANSWER)
			c.Socket.Write(messaging.InfoMessage("Sorry, your answer was not correct. Come back better prepared next time."))
			go c.RemoveFromTrivia()
			itemData, _, err := c.AddItem(&InventorySlot{ItemID: 203001084, Quantity: 1}, -1, false) //false
			if err != nil {
				continue
			}
			c.Socket.Write(*itemData)
		}

		c.TriviaAnswer = 3
	}

	if QuestionRound <= QuestionMaxRounds {
		time.AfterFunc(time.Second*10, func() {
			nextQuestion()
		})
	} else {
		endTriviaEvent()
	}

}
func endTriviaEvent() {

	for _, member := range TriviaParticipants {
		c := member
		c.removeTriviaItems()

		c.Socket.Write(messaging.InfoMessage("Congratulations you answered right to all questions. Here is your reward. Enjoy!"))

		itemData, _, err := c.AddItem(&InventorySlot{ItemID: 203001083, Quantity: 1}, -1, false) //false
		if err != nil {
			continue
		}
		c.Socket.Write(*itemData)
	}
	msg := fmt.Sprintf("Trivia event Round %d ended.", EventRound)
	makeAnnouncement(msg)

	TriviaEventStarted = false
	var part []*Character
	TriviaParticipants = part
	QuestionAsked = nil
	QuestionRound = 0
	EventRound++
	if EventRound <= EventMaxRounds {
		go StartTriviaCounter(60)
	} else {
		EventRound = 1
	}

}

func (c *Character) AnswerTriviaQuestion(answ int) {
	if c.IsTriviaParticipant() {
		if c.TriviaAnswer == 3 {
			c.TriviaAnswer = answ
		} else {
			c.Socket.Write(messaging.InfoMessage("You already answered to this question."))
		}
	}
	c.removeTriviaItems()
}

func (e *QuestionsItem) Create() error {
	return pgsql_DbMap.Insert(e)
}

func (e *QuestionsItem) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *QuestionsItem) Delete() error {
	_, err := pgsql_DbMap.Delete(e)
	return err
}

func (e *QuestionsItem) Update() error {
	_, err := pgsql_DbMap.Update(e)
	return err
}

func getQuestionsItem() error {
	var triviaquestionItems []*QuestionsItem
	query := `select * from data.aso_trivia`

	if _, err := pgsql_DbMap.Select(&triviaquestionItems, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("QuestionsItem: %s", err.Error())
	}

	for _, b := range triviaquestionItems {
		QuestionsList[b.ID] = b
	}

	return nil
}

func RefreshQuestionsItem() error {
	var triviaquestionItems []*QuestionsItem
	query := `select * from data.aso_trivia`

	if _, err := pgsql_DbMap.Select(&triviaquestionItems, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("QuestionsItem: %s", err.Error())
	}

	for _, b := range triviaquestionItems {
		QuestionsList[b.ID] = b
	}

	return nil
}

func (c *Character) AddToTrivia() {
	if !c.IsTriviaParticipant() {
		TriviaParticipants = append(TriviaParticipants, c)
		c.TriviaAnswer = 3
	}
}
func (c *Character) RemoveFromTrivia() {
	var participants []*Character
	for _, member := range TriviaParticipants {
		if member != c {
			participants = append(participants, member)
		}
	}
	TriviaParticipants = participants

}

func (c *Character) IsTriviaParticipant() bool {
	for _, member := range TriviaParticipants {
		if member == c {
			return true
		}
	}
	return false
}

func (c *Character) removeTriviaItems() {
	slot, _, err := c.FindItemInInventory(nil, 17502588) //true
	if err != nil {
	} else if slot != -1 {
		r, err := c.RemoveItem(slot)
		if err != nil {
			log.Print(err)
		}
		c.Socket.Write(r)
	}
	slot, _, err = c.FindItemInInventory(nil, 17502584) //false
	if err != nil {
		log.Println(err)
	} else if slot != -1 {
		r, err := c.RemoveItem(slot)
		if err != nil {
			log.Print(err)
		}
		c.Socket.Write(r)
	}
}

func GetNPCMenu(npcID, textID, index int, actions []int) []byte {
	resp := NPC_MENU
	resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 6)         // npc id
	resp.Insert(utils.IntToBytes(uint64(textID), 4, true), 10)       // text id
	resp.Insert(utils.IntToBytes(uint64(len(actions)), 1, true), 14) // action length

	counter, length := 15, int16(11)
	for i, action := range actions {
		resp.Insert(utils.IntToBytes(uint64(action), 4, true), counter) // action
		counter += 4

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
