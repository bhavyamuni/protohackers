package main

import (
	"log"

	"github.com/BhavyaMuni/protohackers/budgetchat"
	"github.com/BhavyaMuni/protohackers/echo"
	"github.com/BhavyaMuni/protohackers/lineReversal"
	"github.com/BhavyaMuni/protohackers/meanstoanend"
	"github.com/BhavyaMuni/protohackers/mobinthemiddle"
	"github.com/BhavyaMuni/protohackers/primetime"
	"github.com/BhavyaMuni/protohackers/speeddaemon"
	"github.com/BhavyaMuni/protohackers/unusualdatabase"
)

func main() {
	log.Print("Starting servers...")

	go echo.NewEchoServer().Start(":10000")
	go primetime.NewPrimeTimeServer().Start(":10001")
	go meanstoanend.NewMeansToAnEndServer().Start(":10002")
	go budgetchat.NewBudgetChatServer().Start(":10003")
	go unusualdatabase.NewUnusualDatabaseServer().Start(":10004")
	go mobinthemiddle.NewMobInTheMiddleServer().Start(":10005")
	go speeddaemon.NewSpeedDaemonServer().Start(":10006")
	go lineReversal.NewLineReversalServer().Start(":10007")

	select {}
}
