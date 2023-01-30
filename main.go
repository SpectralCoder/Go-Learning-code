package main

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/crud-recipe/handlers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var recipesHandler *handlers.RecipesHandler

func init() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(),
		readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collection := client.Database(os.Getenv(
		"MONGO_DATABASE")).Collection("recipes")
	recipesHandler = handlers.NewRecipesHandler(ctx, collection)
}

// func deleteRecipeHandler(c *gin.Context) {
// 	id := c.Param("id")
// 	index := -1
// 	for i := 0; i < len(recipes); i++ {
// 		if recipes[i].ID == id {
// 			index = i
// 			break
// 		}
// 	}
// 	if index == -1 {
// 		c.JSON(http.StatusNotFound, gin.H{
// 			"error": "recipe not found"})
// 		return
// 	}
// 	recipes = append(recipes[:index], recipes[index+1:]...)
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Recipe has been deleted"})
// }

// func searchRecipesHandler(c *gin.Context) {
// 	tag := c.Query("tag")
// 	listOfRecipes := make([]Recipe, 0)
// 	for i := 0; i < len(recipes); i++ {
// 		found := false
// 		for _, t := range recipes[i].Tags {
// 			if strings.EqualFold(t, tag) {
// 				found = true
// 			}
// 		}
// 		if found {
// 			listOfRecipes = append(listOfRecipes, recipes[i])
// 		}
// 	}
// 	c.JSON(http.StatusOK, listOfRecipes)
// }

func main() {

	router := gin.Default()

	router.POST("/recipes", recipesHandler.NewRecipeHandler)
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.PUT("/recipes/:id", recipesHandler.gg)
	router.Run()

}
