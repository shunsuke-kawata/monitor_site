package main

import (
	"context"
	"fmt"
	"log"
	"monitor_site/model"
	"monitor_site/scraping"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", handleRoot)
	e.GET("/scrape", handleScrape)
	e.GET("/collection/:collection", handleGetCollection)
	e.DELETE("/collection/:collection", handleDeleteCollection)

	e.Logger.Fatal(e.Start(":" + os.Getenv("GOLANG_PORT")))
}

func handleRoot(c echo.Context) error {
	return c.String(http.StatusOK, "root")
}
func handleScrape(c echo.Context) error {
	toLocation := c.QueryParam("to")
	fromLocation := c.QueryParam("from")
	targetDate := c.QueryParam("date")
	if toLocation == "" || fromLocation == "" || targetDate == "" {
		return c.String(http.StatusBadRequest, "to and from query parameters are required")
	}

	targetDateInt, err := strconv.Atoi(targetDate)
	if err != nil {
		return c.String(http.StatusBadRequest, "date must be an integer")
	}

	targetDomain := fmt.Sprintf("https://www.bushikaku.net/search/%s_%s/%s/", fromLocation, toLocation, targetDate)

	// スクレイピング処理を実行
	day, price, fetchedAt, err := scraping.ScrapeBushikaku(targetDomain, targetDateInt)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Scraping error: %v", err))
	}

	// レスポンス構造体を定義
	type ScrapeResponse struct {
		Day   int       `json:"day"`
		Price int       `json:"price"`
		At    time.Time `json:"at"`
	}

	res := ScrapeResponse{
		Day:   day,
		Price: price,
		At:    fetchedAt,
	}
	ctx := context.Background()
	db := &model.MyMongoDB{}
	dbConnectErr := db.Connect(ctx)
	if dbConnectErr != nil {
		log.Fatal(err)
	}
	defer db.Disconnect(ctx)

	collection, dbOperateErr := db.GetOrCreateCollection(ctx, "monitor-db", fmt.Sprintf("%s_%s", fromLocation, toLocation))
	if dbOperateErr != nil {
		log.Fatal(dbOperateErr)
	}
	_, dbInsertErr := collection.InsertOne(ctx, res)
	if dbInsertErr != nil {
		log.Fatal(dbInsertErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Database error: %v", dbInsertErr),
		})
	}

	return c.JSON(http.StatusCreated, res)
}

func handleGetCollection(c echo.Context) error {
	ctx := context.Background()

	db := &model.MyMongoDB{}
	if err := db.Connect(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to connect to MongoDB",
		})
	}
	defer db.Disconnect(ctx)

	collectionName := c.Param("collection")
	documents, err := db.FindAllDocuments(ctx, "monitor-db", collectionName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get or create collection: %v", err),
		})
	}
	for i, doc := range documents {
		fmt.Printf("Document %d: %v\n", i+1, doc)
	}

	return c.JSON(http.StatusOK, documents)
}

func handleDeleteCollection(c echo.Context) error {
	ctx := context.Background()

	db := &model.MyMongoDB{}
	if err := db.Connect(ctx); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to connect to MongoDB",
		})
	}
	defer db.Disconnect(ctx)

	collectionName := c.Param("collection")
	if err := db.DropCollection(ctx, "monitor-db", collectionName); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to delete collection: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Collection %s deleted successfully", collectionName),
	})
}
