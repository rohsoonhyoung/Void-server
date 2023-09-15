package config

import (
	"log"
	"strconv"
)

var Default = &config{
	Database: Database{
		Driver:          "postgres",
		IP:              "localhost",
		Port:            getPort(),
		User:            "postgres",
		Password:        "password", //database password
		Name:            "postgres", //database name
		ConnMaxIdle:     10,
		ConnMaxOpen:     100,
		ConnMaxLifetime: 10,
		Debug:           false,
		SSLMode:         "disable",
	},
	Server: Server{
		IP:   "localhost",
		Port: 4520,
	},
}

func getPort() int {
	sPort := "5432" //database port
	port, err := strconv.ParseInt(sPort, 10, 32)
	if err != nil {
		log.Fatalln(err)
	}

	return int(port)
}
