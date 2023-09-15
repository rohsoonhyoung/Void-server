package auth

import (
	"github.com/bujor2711/Void-server/database"
)

type LauncherHandler struct {
	//password string
	//username string
}

func (lh *LauncherHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	return data, nil
}
