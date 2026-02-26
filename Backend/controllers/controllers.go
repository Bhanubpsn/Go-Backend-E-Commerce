package controllers

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Bhanubpsn/e-commerce-backend/database"
	"github.com/Bhanubpsn/e-commerce-backend/models"
	generate "github.com/Bhanubpsn/e-commerce-backend/token"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var Validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	msg := ""

	if err != nil {
		msg = "Incorrect password"
		valid = false
	}

	return valid, msg
}

// This function will send the email and name to the custom message broker
func SendToBroker(email string, name string) {
	conn, _ := net.Dial("tcp", "localhost:9005")
	defer conn.Close()

	// Send payload
	payload := fmt.Sprintf(`{"email":"%s", "name":"%s"}`, email, name)
	fmt.Fprintln(conn, payload)
}

func Signup() gin.HandlerFunc {
	godotenv.Load()
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationError := Validate.Struct(&user)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationError})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user email already exists"})
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})

		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user phone already exists"})
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refreshToken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		user.Token = &token
		user.Refresh_Token = &refreshToken
		user.UserCart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0)

		_, inserter := UserCollection.InsertOne(ctx, user)
		if inserter != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User did not get created"})
			return
		}
		defer cancel()
		SendToBroker(*user.Email, *user.First_Name)
		c.JSON(http.StatusCreated, "Successfully signed in: token: "+token)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var founduser models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password Incorret"})
			return
		}

		PasswordIsValid, msg := VerifyPassword(*user.Password, *founduser.Password)
		defer cancel()

		if !PasswordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}

		token, refreshToken, _ := generate.TokenGenerator(*founduser.Email, *founduser.First_Name, *founduser.Last_Name, founduser.User_ID)
		defer cancel()

		generate.UpdateAllTokens(token, refreshToken, founduser.User_ID)
		c.JSON(http.StatusFound, founduser)
	}
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var products models.Product
		defer cancel()

		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		products.Product_ID = primitive.NewObjectID()
		_, err := ProductCollection.InsertOne(ctx, products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Not addede the product"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, "Successfully added")
	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productList []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		secondaryOpts := options.Collection().SetReadPreference(readpref.SecondaryPreferred())

		// For secondary reads
		searchCollection, err := ProductCollection.Clone(secondaryOpts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Secondary read preference error: " + err.Error()})
			return
		}

		cursor, err := searchCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "Something went wrong")
			return
		}

		err = cursor.All(ctx, &productList)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		defer cursor.Close(ctx)
		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "Invalid")
			return
		}

		defer cancel()
		c.IndentedJSON(200, productList)
	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var SearchProducts []models.Product

		// Get both parameters from the URL: /users/search?name=iphone&category=electronics
		nameQuery := c.Query("name")
		categoryQuery := c.Query("category")

		if nameQuery == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query name is required"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{
			"$text": bson.M{"$search": nameQuery},
		}
		if categoryQuery != "" {
			filter["category"] = categoryQuery
		}

		// Pagation: limit to 20 results
		findOptions := options.Find().SetLimit(20)

		// Use a secondary read preference to distribute read load
		secondaryRead := options.Collection().SetReadPreference(readpref.SecondaryPreferred())

		// Hopefully the clone does not make a copy of the entire collection, but just a shallow one
		secondaryCol, err := ProductCollection.Clone(secondaryRead)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Config error"})
			return
		}

		cursor, err := secondaryCol.Find(ctx, filter, findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &SearchProducts); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Decode error"})
			return
		}

		c.JSON(http.StatusOK, SearchProducts)
	}
}
