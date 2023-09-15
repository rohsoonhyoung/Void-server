package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/bujor2711/Void-server/ai"
	"github.com/bujor2711/Void-server/config"
	"github.com/bujor2711/Void-server/database"
	"github.com/bujor2711/Void-server/events"
	_ "github.com/bujor2711/Void-server/factory"
	"github.com/bujor2711/Void-server/logging"
	"github.com/bujor2711/Void-server/nats"
	"github.com/bujor2711/Void-server/redis"
	"github.com/go-co-op/gocron"
)

var (
	logger    = logging.Logger
	CacheFile = "cache.json"
)

func initDB_PostgreSQL() {
	for {
		err := database.InitPostgreSQL()
		if err == nil {
			log.Print("Connected to PostgreSQL database...")
			return
		}
		log.Print(fmt.Sprintf("PostgreSQL Database connection error: %+v, waiting 30 sec...", err))
		time.Sleep(time.Duration(30) * time.Second)
	}
}

func initRedis() {
	for {
		err := redis.InitRedis()
		if err != nil {
			log.Printf("Redis connection error: %+v, waiting 30 sec...", err)
			time.Sleep(time.Duration(30) * time.Second)
			continue
		}

		if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
			log.Printf("Connected to redis...")
			go logger.StartLogging()
		}

		return
	}
}
func StartLogging() {
	fi, err := os.OpenFile("Log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //log file
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(fi)
}

func startServer() {
	cfg := config.Default
	port := cfg.Server.Port
	listen, err := net.Listen("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Printf("Socket listen port %d failed,%s", port, err)
		os.Exit(1)
	}
	defer listen.Close()
	log.Printf("Begin listen port: %d", port)
	//StartLogging()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}

		ws := database.Socket{Conn: conn}
		go ws.Read()

	}

}

func cronHandler() {
	var err error
	s := gocron.NewScheduler(time.Local)
	task1 := func() {
		err = database.RefreshAIDs()
		err = database.RefreshYingYangKeys()
	}
	_, err = s.Every("24h").Do(task1)

	s.StartAsync()
	if err != nil {
		fmt.Print(err)
	}
}
func main() {

	var err error

	log.Print("-----------------Initialize pgsql-------------------------------")
	initDB_PostgreSQL()
	log.Print("--------------------------------------------------------------")

	cronHandler()

	ai.Init()
	//ai.InitBabyPets()
	ai.InitHouseItems()
	go database.HandleClanBuffs()
	//go database.StartLoto()
	go database.UnbanUsers()
	//go database.InitDiscordBot()
	go database.DeleteInexistentItems()
	//go database.FactionWarSchedule()
	go database.DeleteUnusedStats()

	go events.CiEventSchedule()

	s := nats.RunServer(nil)
	defer s.Shutdown()
	c, err := nats.ConnectSelf(nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()
	go database.EpochHandler()
	startServer()

}
