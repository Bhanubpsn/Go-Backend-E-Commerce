package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"github.com/Bhanubpsn/e-commerce-backend/controllers"
	"github.com/Bhanubpsn/e-commerce-backend/database"
	"github.com/Bhanubpsn/e-commerce-backend/middleware"
	"github.com/Bhanubpsn/e-commerce-backend/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := controllers.NewApplication(
		database.ProductData(database.Client, "Products"), 
		database.UserData(database.Client, "Users"),
	)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf(
			"%s - [%s] \"%s %s %d %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))

	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoveItem())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())

	log.Fatal(router.Run(":" + port))
}