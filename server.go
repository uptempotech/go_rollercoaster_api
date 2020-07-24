package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/joho/godotenv"
	"github.com/uptempotech/go_rollercoaster_api/global"
)

var coastersCollection mongo.Collection

// Coaster struct defines a coaster
type Coaster struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	ID           string `json:"id"`
	InPark       string `json:"inPark"`
	Height       int    `json:"height"`
}

type coasterHandlers struct{}

func (h *coasterHandlers) coasters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

func (h *coasterHandlers) get(w http.ResponseWriter, r *http.Request) {
	var coasters []Coaster
	ctx, cancel := global.NewDBContext(5 * time.Second)
	defer cancel()

	cursor, err := coastersCollection.Find(ctx, bson.D{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		data := &global.Coaster{}

		err = cursor.Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		newCoaster := Coaster{
			Name:         data.Name,
			Manufacturer: data.Manufacturer,
			ID:           data.CoasterID,
			InPark:       data.InPark,
			Height:       data.Height,
		}

		coasters = append(coasters, newCoaster)
	}

	jsonBytes, err := json.Marshal(coasters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *coasterHandlers) getCoaster(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ctx, cancel := global.NewDBContext(5 * time.Second)
	defer cancel()

	var data global.Coaster
	filter := bson.M{"coaster_id": parts[2]}
	coastersCollection.FindOne(ctx, filter).Decode(&data)
	if data == global.NilCoaster {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	coaster := Coaster{
		Name:         data.Name,
		Manufacturer: data.Manufacturer,
		ID:           data.CoasterID,
		InPark:       data.InPark,
		Height:       data.Height,
	}

	jsonBytes, err := json.Marshal(coaster)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *coasterHandlers) post(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("need content-type 'application/json', but got '%s'", ct)))
		return
	}

	var coaster Coaster
	err = json.Unmarshal(bodyBytes, &coaster)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	id := primitive.NewObjectID()
	coaster.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	mCoster := &global.Coaster{
		ID:           id,
		Name:         coaster.Name,
		Manufacturer: coaster.Manufacturer,
		CoasterID:    coaster.ID,
		InPark:       coaster.InPark,
		Height:       coaster.Height,
	}

	ctx, cancel := global.NewDBContext(5 * time.Second)
	defer cancel()

	_, err = coastersCollection.InsertOne(ctx, mCoster)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func newCoasterHandlers() *coasterHandlers {
	return &coasterHandlers{}
}

type adminPortal struct {
	password string
}

func newAdminPortal() *adminPortal {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		panic("required env var ADMIN_PASSWORD not set")
	}

	return &adminPortal{password: password}
}

func (a adminPortal) handler(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - unauthorized"))
		return
	}

	w.Write([]byte("<html><h1>Super secret admin portal</h1></html>"))
}

func main() {
	coastersCollection = *global.DB.Collection("roller_coasters")

	//admin := newAdminPortal()

	coasterHandlers := newCoasterHandlers()
	http.HandleFunc("/coasters", coasterHandlers.coasters)
	http.HandleFunc("/coasters/", coasterHandlers.getCoaster)
	//http.HandleFunc("/admin", admin.handler)

	err := http.ListenAndServe(":8082", nil)
	if err != nil {
		panic(err)
	}
}
