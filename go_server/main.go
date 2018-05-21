package main

import (
	"gopkg.in/mgo.v2"
	"flag"
	"os"
	"gopkg.in/mgo.v2/bson"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
	"os/signal"
	"syscall"
	"log"
	"strings"
)

var gifs Results

var db *mgo.Database

const (
	COLLECTION = "gifs"
)

var port = flag.String("port", "", "port to run the server (Required)")

// main function to boot up everything
func main() {
	flag.Parse()
	if *port == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	session := connectToDB(DB{connectionPoint: "localhost"})
	defer session.Close()
	db = session.DB("testDB")

	//Remove old data from DB
	db.C(COLLECTION).RemoveAll(bson.M{})

	insertInitialData()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello page"))
	})

	router.Route("/gifs", func(r chi.Router) {
		r.Get("/", getGifs)
		r.Post("/", createGifs)

		r.Route("/{gifID}", func(r chi.Router) {
			r.Use(Ctx)
			r.Get("/", getGif)
			r.Delete("/", deleteGif)
		})
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	go func() {
		<-sigs
		saveData()
		os.Exit(0)
	}()

	log.Fatal(http.ListenAndServe(strings.Join([]string{"", *port}, ":"), router))

}
