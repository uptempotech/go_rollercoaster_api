package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/joho/godotenv"
	"github.com/uptempotech/go_rollercoaster_api/grpc_client/proto"
)

var client proto.CoasterServiceClient

// Coaster struct defines a coaster
type Coaster struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	ID           string `json:"id"`
	InPark       string `json:"inPark"`
	Height       int32  `json:"height"`
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

	data := &proto.GetCoastersRequest{
		Empty: "",
	}

	res, err := client.GetCoasters(context.Background(), data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for _, coaster := range res.Coasters {
		newCoaster := Coaster{
			Name:         coaster.Name,
			Manufacturer: coaster.Manufacturer,
			ID:           coaster.CoasterID,
			InPark:       coaster.InPark,
			Height:       coaster.Height,
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

	req := &proto.GetCoasterRequest{
		CoasterID: parts[2],
	}

	res, err := client.GetCoaster(context.Background(), req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	coaster := Coaster{
		Name:         res.Coaster.Name,
		Manufacturer: res.Coaster.Manufacturer,
		ID:           res.Coaster.CoasterID,
		InPark:       res.Coaster.InPark,
		Height:       res.Coaster.Height,
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

	coaster.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	_, err = client.AddNewCoaster(context.Background(), &proto.AddNewCoasterRequest{
		Coaster: &proto.RollerCoaster{
			Name:         coaster.Name,
			Manufacturer: coaster.Manufacturer,
			CoasterID:    coaster.ID,
			InPark:       coaster.InPark,
			Height:       coaster.Height,
		},
	})

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
	conn, err := grpc.Dial("localhost:5000", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client = proto.NewCoasterServiceClient(conn)

	//admin := newAdminPortal()

	coasterHandlers := newCoasterHandlers()
	http.HandleFunc("/coasters", coasterHandlers.coasters)
	http.HandleFunc("/coasters/", coasterHandlers.getCoaster)
	//http.HandleFunc("/admin", admin.handler)

	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		panic(err)
	}
}
