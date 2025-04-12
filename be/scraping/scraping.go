package scraping

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func getDay(date int) int {
	dateStr := strconv.Itoa(date)
	dayStr := dateStr[len(dateStr)-2:]
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		log.Println("Error converting day to integer:", err)
		return 0
	}
	return day
}

func ScrapeBushikaku(path string, date int) (int, int, time.Time, error) {
	day := getDay(date)

	var scrapeErr error
	var price int
	var fetchedAt time.Time

	c := colly.NewCollector(
		colly.CacheDir("./cache"),
		colly.UserAgent("Sample-Scraper"),
	)

	c.OnHTML(".SearchLowestPriceCalendar_day-button__qHW2J", func(e *colly.HTMLElement) {
		dayText := strings.TrimSpace(e.DOM.Find("span.SearchLowestPriceCalendar_day-number__ZBLQq").Text())

		if dayText == strconv.Itoa(day) {
			priceText := strings.TrimSpace(e.DOM.Find("span.SearchLowestPriceCalendar_amount__NQXRS span").Text())
			priceText = strings.Replace(priceText, "円", "", -1)
			priceText = strings.Replace(priceText, ",", "", -1)

			var err error
			price, err = strconv.Atoi(priceText)
			if err != nil {
				scrapeErr = fmt.Errorf("金額の変換に失敗しました: %v", err)
				return
			}

			fetchedAt = time.Now()
		}
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Error:", err)
		scrapeErr = err
	})

	err := c.Visit(path)
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("visit error: %w", err)
	}

	if scrapeErr != nil {
		return 0, 0, time.Time{}, scrapeErr
	}

	if price == 0 {
		return 0, 0, time.Time{}, fmt.Errorf("日付 %d に一致する金額が見つかりませんでした", date)
	}

	return day, price, fetchedAt, nil
}
