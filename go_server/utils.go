package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"os"
	"fmt"
	"log"
	"gopkg.in/mgo.v2/bson"
	"github.com/go-chi/chi"
	"context"
)

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
	body, err := ioutil.ReadFile("data_gifs")
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

func check(e error) {
	if e != nil {
		//TODO: Add better logging
		log.Fatal(e)
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

