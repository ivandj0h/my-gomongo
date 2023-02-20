package main

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

type Item struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name,omitempty"`
	Price int                `json:"price,omitempty" bson:"price,omitempty"`
}

func main() {

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	itemsCollection := client.Database("mygomongo").Collection("items")

	http.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Get all items
			cur, err := itemsCollection.Find(context.Background(), primitive.D{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cur.Close(context.Background())
			var items []Item
			for cur.Next(context.Background()) {
				var item Item
				if err := cur.Decode(&item); err != nil {
					log.Println(err)
					continue
				}
				items = append(items, item)
			}
			if err := cur.Err(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(items)

		case http.MethodPost:
			// Create a new item
			var item Item
			if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if item.Name == "" || item.Price == 0 {
				http.Error(w, "Name and price are required", http.StatusBadRequest)
				return
			}

			result, err := itemsCollection.InsertOne(context.Background(), item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(result.InsertedID)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Get a single item
			id := r.URL.Path[len("/items/"):]
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var item Item
			err = itemsCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&item)
			if err != nil {
				http.Error(w, "Item not found", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(item)

		case http.MethodPut:
			// Update an item
			id := r.URL.Path[len("/items/"):]
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			var item Item
			if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			update := bson.M{
				"$set": bson.M{
					"name":  item.Name,
					"price": item.Price,
				},
			}

			_, err = itemsCollection.UpdateOne(context.Background(), bson.M{"_id": objectID}, update)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			// Delete an item
			id := r.URL.Path[len("/items/"):]
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			_, err = itemsCollection.DeleteOne(context.Background(), bson.M{"_id": objectID})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start the HTTP server
	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
