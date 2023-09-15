package database

import (
	"fmt"
	"math"
	"time"

	"github.com/bujor2711/Void-server/messaging"
	"github.com/bujor2711/Void-server/utils"
)

var (
	isActive        = false
	participants    []*Participant
	lotoPrice       = uint64(500)
	minParticipants = 3
	countDownTimer  = 600 //seconds
)

type Participant struct {
	character *Character
	number    int
}

func StartLoto() {
	time.AfterFunc(time.Hour*3, func() {
		msg := "Lottery event has started. Buy loto ticket by typing /loto. Price : 500nC"
		makeAnnouncement(msg)
		go StartLoto()
		go CountLoto(countDownTimer)
	})
}
func CountLoto(cd int) {
	isActive = true
	if cd >= 120 {
		checkMembersInFactionWarMap()
		msg := fmt.Sprintf("Lottery number will be extracted in %d minutes.", cd/60)
		makeAnnouncement(msg)
		time.AfterFunc(time.Second*60, func() {
			CountLoto(cd - 60)
		})
	} else if cd > 0 {
		checkMembersInFactionWarMap()
		msg := fmt.Sprintf("Lottery number will be extracted in %d seconds.", cd)
		makeAnnouncement(msg)
		time.AfterFunc(time.Second*10, func() {
			CountLoto(cd - 10)
		})
	}
	if cd <= 0 {
		endLoto()
		participants = nil
		isActive = false
	}
}

func endLoto() {
	if len(participants) == 0 {
		return
	}
	if len(participants) < minParticipants {
		for _, participant := range participants {
			user, err := FindUserByID(participant.character.UserID)
			if err != nil {
				return
			} else if user == nil {
				return
			}
			user.NCash += lotoPrice
			user.Update()
			msg := "Lottery event canceled because not enough participants to the event."
			makeAnnouncement(msg)
			msg = "Not enough participants to lottery. The ticket value has been restored to your account."
			participant.character.Socket.Write(messaging.InfoMessage(msg))
		}
	} else {
		rand := int(utils.RandInt(0, 100))
		difference := 100
		winner := &Participant{}
		for _, participant := range participants {
			if int(math.Abs(float64(rand)-float64(participant.number))) < difference {
				difference = int(math.Abs(float64(rand) - float64(participant.number)))
				winner = participant
			}
		}
		user, err := FindUserByID(winner.character.UserID)
		if err != nil {
			return
		} else if user == nil {
			return
		}
		winnings := lotoPrice * uint64(len(participants))
		user.NCash += winnings
		user.Update()
		msg := fmt.Sprintf("Lottery: %s has aquired %d ncash by winning the Lottery.", winner.character.Name, winnings)
		makeAnnouncement(msg)
		msg = fmt.Sprintf("Extracted number : %d. Congratulations, you won the Lottery with number : %d. %d Ncash has been added to your account.", rand, winner.number, winnings)
		winner.character.Socket.Write(messaging.InfoMessage(msg))
		for _, participant := range participants {
			if participant != winner {
				msg = fmt.Sprintf("Extracted number : %d. You chose %d and lost :( Good luck next time.", rand, participant.number)
				participant.character.Socket.Write(messaging.InfoMessage(msg))
			}
		}
	}
}
func AddPlayer(s *Socket, number int) {
	if !isActive {
		msg := "Loto event is not active at this moment."
		s.Write(messaging.InfoMessage(msg))
		return
	}
	if number < 0 || number > 100 {
		msg := "You must choose a number between 0 and 100!"
		s.Write(messaging.InfoMessage(msg))
		return
	}
	for _, participant := range participants {
		if participant.character == s.Character {
			msg := "You already have a ticket. Wait for loto number extraction. Good Luck!"
			s.Write(messaging.InfoMessage(msg))
			return
		}
		if participant.number == number {
			msg := "This number was already picked, choose another number."
			s.Write(messaging.InfoMessage(msg))
			return
		}
	}
	if s.User.NCash < lotoPrice {
		msg := fmt.Sprintf("You don't have enough cash to buy loto ticket. Price : %d", int(lotoPrice))
		s.Write(messaging.InfoMessage(msg))
		return
	} else {
		s.User.NCash -= lotoPrice
		s.User.Update()
		participant := &Participant{
			character: s.Character,
			number:    number,
		}
		participants = append(participants, participant)

		for _, participant := range participants {
			msg := fmt.Sprintf("%s bought a lotery ticket with number :%d", s.Character.Name, number)
			participant.character.Socket.Write(messaging.InfoMessage(msg))
		}
	}
}
