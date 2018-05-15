package main

import (
	"encoding/json"
	"log"
	"net/http"
	"io/ioutil"
	"os"
	"gopkg.in/mgo.v2"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"os/signal"
	"syscall"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"flag"
	"strings"
	"context"
)

type Element struct {
	Id    string `bson:"id" json:"id"`
	Url   string `bson:"url" json:"url"`
	Title string `bson:"title" json:"title"`
}

type Results struct {
	Data []Element `json:"data"`
}

var gifs Results

func Ctx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gifID := chi.URLParam(r, "gifID")

		var element Element
		err := db.C(COLLECTION).Find(bson.M{"id": gifID}).One(&element)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "gif", element)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Get items
func getGifs(w http.ResponseWriter, r *http.Request) {
	var elements []Element
	err := db.C(COLLECTION).Find(bson.M{}).All(&elements)
	check(err)
	json.NewEncoder(w).Encode(Results{elements})
}

// Get an item
func getGif(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	element, true := ctx.Value("gif").(Element)
	log.Println(element)
	if !true {
		http.Error(w, http.StatusText(422), 422)
		return
	}
	json.NewEncoder(w).Encode(element)
}

// Create a new item
func createGifs(w http.ResponseWriter, r *http.Request) {
	var element Element
	err := json.NewDecoder(r.Body).Decode(&element)
	check(err)
	if element == (Element{}) {
		json.NewEncoder(w).Encode("Please use correct format")
		return
	}
	db.C(COLLECTION).Insert(&element)
	json.NewEncoder(w).Encode(element)
}

// Delete an item
func deleteGif(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	element, true := ctx.Value("gif").(Element)
	if !true {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	err := db.C(COLLECTION).Remove(element)
	check(err)
	json.NewEncoder(w).Encode("Element was deleted")
}

func loadData() Results {
	var s = new(Results)
	body, err := ioutil.ReadFile("data_gif")
	if err != nil {
		os.Exit(1)
	}
	json.Unmarshal(body, &s)
	return *s
}

func insertInitialData() {
	gifs = loadData()
	for _, element := range gifs.Data {
		err := db.C(COLLECTION).Insert(element)
		check(err)
	}
}

func connectToDB() *mgo.Session {
	session, err := mgo.Dial("localhost")
	if err != nil {
		fmt.Println("Hello it is an error occured:", err)
		os.Exit(1)
	}
	session.SetMode(mgo.Monotonic, true)
	// Error check on every access
	session.SetSafe(&mgo.Safe{})

	return session
}

func check(e error) {
	if e != nil {
		//TODO: Add better logging
		log.Fatal(e)
		os.Exit(0)
	}
}

func saveData() {
	var elements []Element
	err := db.C(COLLECTION).Find(bson.M{}).All(&elements)
	check(err)
	res := Results{elements}

	f, err := os.Create("data_gifs")
	check(err)
	fmt.Println("Store data.")
	b, err := json.Marshal(res)
	f.Write(b)

	defer f.Close()
}

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

	session := connectToDB()
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
