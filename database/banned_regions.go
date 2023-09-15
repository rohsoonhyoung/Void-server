package database

import (
	"database/sql"
	"fmt"
)

var (
	BannedRegions = make(map[int]*BannedRegion)
)

type BannedRegion struct {
	Id     int    `db:"id"`
	Region string `db:"region"`
}

func (b *BannedRegion) Update() error {
	_, err := pgsql_DbMap.Update(b)
	return err
}

func GetBannedBannedRegions() error {
	var ips []*BannedRegion
	query := `select * from hops.banned_regions`

	if _, err := pgsql_DbMap.Select(&ips, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("GetBannedBannedRegions: %s", err.Error())
	}

	for _, cr := range ips {
		BannedRegions[cr.Id] = cr
	}
	return nil
}
