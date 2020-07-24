package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/uptempotech/go_rollercoaster_api/grpc_server/global"
	"github.com/uptempotech/go_rollercoaster_api/grpc_server/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type coasterServer struct{}

var coastersCollection mongo.Collection

func (server coasterServer) AddNewCoaster(_ context.Context, in *proto.AddNewCoasterRequest) (*proto.AddNewCoasterResponse, error) {
	newCoaster := global.NewCoaster()
	newCoaster.Name = in.Coaster.GetName()
	newCoaster.Manufacturer = in.Coaster.GetManufacturer()
	newCoaster.CoasterID = in.Coaster.GetCoasterID()
	newCoaster.InPark = in.Coaster.GetInPark()
	newCoaster.Height = in.Coaster.GetHeight()

	ctx, cancel := global.NewDBContext(5 * time.Second)
	defer cancel()
	_, err := coastersCollection.InsertOne(ctx, newCoaster)
	if err != nil {
		return &proto.AddNewCoasterResponse{Result: "Failed to add new coaster", Success: false}, status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error: %v", err))
	}

	return &proto.AddNewCoasterResponse{Result: "Added New Coaster", Success: true}, nil
}

func (server coasterServer) GetCoasters(_ context.Context, in *proto.GetCoastersRequest) (*proto.GetCoastersResponse, error) {
	ctx, cancel := global.NewDBContext(5 * time.Second)
	defer cancel()

	coasters := []*proto.RollerCoaster{}

	cursor, err := coastersCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error: %v", err))
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		data := &global.Coaster{}

		err := cursor.Decode(&data)
		if err != nil {
			return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("Could not decode data: %v", err))
		}

		newCoaster := &proto.RollerCoaster{
			Name:         data.Name,
			Manufacturer: data.Manufacturer,
			CoasterID:    data.CoasterID,
			InPark:       data.InPark,
			Height:       data.Height,
		}

		coasters = append(coasters, newCoaster)
	}

	return &proto.GetCoastersResponse{Coasters: coasters}, nil
}

func (server coasterServer) GetCoaster(_ context.Context, in *proto.GetCoasterRequest) (*proto.GetCoasterResponse, error) {
	filter := in.GetCoasterID()

	ctx, cancel := global.NewDBContext(5 * time.Second)
	defer cancel()

	var data global.Coaster

	coastersCollection.FindOne(ctx, bson.M{"coaster_id": filter}).Decode(&data)
	coaster := &proto.RollerCoaster{
		Name:         data.Name,
		Manufacturer: data.Manufacturer,
		CoasterID:    data.CoasterID,
		InPark:       data.InPark,
		Height:       data.Height,
	}

	return &proto.GetCoasterResponse{Coaster: coaster}, nil
}

func main() {
	coastersCollection = *global.DB.Collection("roller_coasters")

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	proto.RegisterCoasterServiceServer(grpcServer, coasterServer{})
	log.Println("Starting gRPC server on port 5000")

	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatal("Error creating listener: ", err.Error())
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("Error setting up gRPC server.")
	}
}
