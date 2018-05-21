package main

import (
	"gopkg.in/mgo.v2"
	"fmt"
	"os"
)

type Element struct {
	Id    string `bson:"id" json:"id"`
	Url   string `bson:"url" json:"url"`
	Title string `bson:"title" json:"title"`
}

type Results struct {
	Data []Element `json:"data"`
}

type session *mgo.Session


type DBase interface {
	Connect() (session, error)
}


type DB struct{
	connectionPoint string
}

func(dataBase DB) Connect() (session, error){
	dbSession, err := mgo.Dial(dataBase.connectionPoint)
	dbSession.SetMode(mgo.Monotonic, true)
	// Error check on every access
	dbSession.SetSafe(&mgo.Safe{})

	return dbSession, err
}

func connectToDB(conn DBase) *mgo.Session {
	session, err := conn.Connect()
	if err != nil {
		fmt.Println("Hello it is an error occured:", err)
		os.Exit(1)
	}
	return session
}
