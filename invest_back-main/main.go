package main

import (
	"investtrack-backend/database"
	"investtrack-backend/handlers"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	database.Init()

	// Ensure uploads directory exists
	os.MkdirAll("./uploads", 0755)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// Serve uploaded files as static assets
	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		api.GET("/healthz", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "service": "investtrack-go-api"})
		})

		// File upload
		api.POST("/upload", handlers.UploadImage)

		api.GET("/objects", handlers.GetObjects)
		api.GET("/objects/top", handlers.GetTopObjects)
		api.GET("/objects/:id", handlers.GetObject)
		api.POST("/objects", handlers.CreateObject)
		api.PUT("/objects/:id", handlers.UpdateObject)
		api.DELETE("/objects/:id", handlers.DeleteObject)

		api.GET("/owners", handlers.GetOwners)
		api.GET("/owners/:id", handlers.GetOwner)
		api.POST("/owners", handlers.CreateOwner)
		api.PUT("/owners/:id", handlers.UpdateOwner)
		api.DELETE("/owners/:id", handlers.DeleteOwner)

		api.GET("/investors", handlers.GetInvestors)
		api.GET("/investors/:id", handlers.GetInvestor)
		api.POST("/investors", handlers.CreateInvestor)
		api.PUT("/investors/:id", handlers.UpdateInvestor)
		api.DELETE("/investors/:id", handlers.DeleteInvestor)

		api.GET("/object-images", handlers.GetObjectImages)
		api.POST("/object-images", handlers.CreateObjectImage)
		api.DELETE("/object-images/:id", handlers.DeleteObjectImage)

		api.GET("/comments", handlers.GetComments)
		api.POST("/comments", handlers.CreateComment)
		api.DELETE("/comments/:id", handlers.DeleteComment)

		api.GET("/statistics", handlers.GetStatistics)
		api.POST("/statistics", handlers.CreateStatistic)
		api.PUT("/statistics/:id", handlers.UpdateStatistic)
		api.DELETE("/statistics/:id", handlers.DeleteStatistic)

		api.GET("/faqs", handlers.GetFAQs)
		api.POST("/faqs", handlers.CreateFAQ)
		api.PUT("/faqs/:id", handlers.UpdateFAQ)
		api.DELETE("/faqs/:id", handlers.DeleteFAQ)

		api.GET("/dashboard/summary", handlers.GetDashboardSummary)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("InvesTrack Go API starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
