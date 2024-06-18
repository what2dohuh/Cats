package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	Client *mongo.Client
}

func newServer(c *mongo.Client) *Server {
	return &Server{
		Client: c,
	}
}

func (s *Server) HandleGetReq(w http.ResponseWriter, r *http.Request) {
	coll := s.Client.Database("Cats").Collection("facts")
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	result := []bson.M{}
	if err := cursor.All(context.TODO(), &result); err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

type workerCat struct {
	Client *mongo.Client
}

func newworkerCat(c *mongo.Client) *workerCat {
	return &workerCat{
		Client: c,
	}
}

func (cw *workerCat) start() error {
	coll := cw.Client.Database("Cats").Collection("facts")
	ticker := time.NewTicker(2 * time.Second)
	for {
		resp, err := http.Get("http://catfact.ninja/fact")
		if err != nil {
			log.Fatal(err)
		}
		var catfact bson.M
		if err = json.NewDecoder(resp.Body).Decode(&catfact); err != nil {
			log.Fatal(err)
		}
		if _, err = coll.InsertOne(context.TODO(), catfact); err != nil {
			log.Fatal(err)
		}
		<-ticker.C

	}

}
func main() {
	godotenv.Load()
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGO")))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(client)
	worker := newworkerCat(client)
	go worker.start()

	server := newServer(client)
	http.HandleFunc("/cats", server.HandleGetReq)
	http.ListenAndServe(os.Getenv("PORT"), nil)
}
