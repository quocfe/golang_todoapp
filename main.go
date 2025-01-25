package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	// Initialize a new Fiber app
	app := fiber.New()

	// cors
	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: "http://localhost:5173",
	// 	AllowHeaders: "Origin,Content-Type,Accept",
	// }))
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file", err)
		}
	}
	MONGODB_URI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB")

	collection = client.Database("todoAppGoLang").Collection("todos")

	app.Get("/api/todos", func(c *fiber.Ctx) error {
		var todos []Todo
		cursor, err := collection.Find(context.Background(), bson.M{})

		if err != nil {
			return err
		}
		// cursor là con trỏ, trỏ đến tập hợp các tài liệu
		defer cursor.Close(context.Background()) // Đảm bảo đóng con trỏ sau khi dùng xong

		for cursor.Next(context.Background()) {
			var todo Todo
			if err := cursor.Decode(&todo); err != nil {
				return err
			}
			todos = append(todos, todo)
		}

		return c.Status(fiber.StatusOK).JSON(todos)
	})

	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := new(Todo)

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		if todo.Body == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Todo body cannot be empty"})
		}

		insertResult, err := collection.InsertOne(context.Background(), todo)

		if err != nil {
			return err
		}

		todo.ID = insertResult.InsertedID.(primitive.ObjectID)

		return c.Status(fiber.StatusCreated).JSON(todo)
	})

	app.Patch("/api/todo/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		objectID, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
		}

		filter := bson.M{"_id": objectID}
		update := bson.M{"$set": bson.M{"completed": true}}

		_, err = collection.UpdateOne(context.Background(), filter, update)

		if err != nil {
			return err
		}

		return c.Status(200).JSON(fiber.Map{"success": true})
	})

	app.Delete("/api/todo/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		objectID, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
		}

		filter := bson.M{"_id": objectID}
		_, err = collection.DeleteOne(context.Background(), filter)

		if err != nil {
			return err
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
	})

	// Start the server on port 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/dist")
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}
