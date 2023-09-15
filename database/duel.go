package database

import "github.com/bujor2711/Void-server/utils"

type Duel struct {
	EnemyID    int
	Coordinate utils.Location
	Started    bool
}
