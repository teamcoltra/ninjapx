package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// Structs for JSON responses, extended to include PageURL
type TodayStats struct {
	StatsByPageURL []PageStat `json:"statsByPageURL"`
}

type PageStat struct {
	PageURL      string    `json:"pageURL"`
	PageViews    int       `json:"pageViews"`
	UniqueVisits int       `json:"uniqueVisits"`
	GeoStats     []GeoStat `json:"geoStats"`
}

type GeoStat struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
	Views   int    `json:"views"`
}

type HistoricalStats struct {
	StatsByPageURL []HistoricalPageStat `json:"statsByPageURL"`
}

type HistoricalPageStat struct {
	PageURL          string         `json:"pageURL"`
	PageViews        int            `json:"pageViews"`
	Bounces          int            `json:"bounces"`
	Referrals        []ReferralStat `json:"referrals"`
	GeolocationStats []GeoStat      `json:"geolocationStats"`
}

type ReferralStat struct {
	ReferralURL string `json:"referralURL"`
	Count       int    `json:"count"`
}

// Handler for today's stats with PageURL grouping and filtering
func serveTodayStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pageURL := r.URL.Query().Get("pageURL") // For filtering by PageURL if provided

	var todayStats TodayStats
	var pageStats []PageStat

	// Modify SQL to include PageURL grouping and optional filtering
	pageURLCondition := ""
	if pageURL != "" {
		pageURLCondition = fmt.Sprintf("AND PageURL = '%s'", pageURL)
	}

	todayQuery := fmt.Sprintf(`
        SELECT PageURL, COUNT(*) AS PageViews, COUNT(DISTINCT IPAddress) AS UniqueVisits
        FROM log
        WHERE DATE(Timestamp) = DATE('now') %s
        GROUP BY PageURL;
    `, pageURLCondition)

	rows, err := db.Query(todayQuery)
	if err != nil {
		http.Error(w, "Error querying today's stats", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stat PageStat
		err := rows.Scan(&stat.PageURL, &stat.PageViews, &stat.UniqueVisits)
		if err != nil {
			http.Error(w, "Error reading today's stats", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		// Fetch and append GeoStats here for each PageURL
		geoQuery := fmt.Sprintf(`
            SELECT GeoCity, GeoState, GeoCountry, COUNT(*) AS Views
            FROM log
            WHERE DATE(Timestamp) = DATE('now') AND PageURL = '%s'
            GROUP BY GeoCity, GeoState, GeoCountry;
        `, stat.PageURL)

		geoRows, err := db.Query(geoQuery)
		if err != nil {
			http.Error(w, "Error querying geolocation stats", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		for geoRows.Next() {
			var geoStat GeoStat
			err := geoRows.Scan(&geoStat.City, &geoStat.State, &geoStat.Country, &geoStat.Views)
			if err != nil {
				http.Error(w, "Error reading geolocation stats", http.StatusInternalServerError)
				log.Println(err)
				return
			}
			stat.GeoStats = append(stat.GeoStats, geoStat)
		}
		geoRows.Close()

		pageStats = append(pageStats, stat)
	}

	todayStats.StatsByPageURL = pageStats
	json.NewEncoder(w).Encode(todayStats)
}

// Handler for historical stats with PageURL grouping and filtering
func serveHistoricalStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	days, err := strconv.Atoi(r.URL.Query().Get("days"))
	if err != nil || days <= 0 {
		days = 30 // Default to last 30 days if not specified or invalid
	}
	pageURL := r.URL.Query().Get("pageURL") // For filtering by PageURL if provided

	var historicalStats HistoricalStats
	var statsByPageURL []HistoricalPageStat

	pageURLCondition := ""
	if pageURL != "" {
		pageURLCondition = fmt.Sprintf("AND PageURL = '%s'", pageURL)
	}

	// Modify historical stats SQL to include PageURL grouping and optional filtering
	historicalQuery := fmt.Sprintf(`
        SELECT PageURL, SUM(PageViews) AS PageViews, SUM(Bounces) AS Bounces
        FROM aggregated_analytics
        WHERE Datestamp >= DATE('now', '-%d days') %s
        GROUP BY PageURL;
    `, days, pageURLCondition)

	rows, err := db.Query(historicalQuery)
	if err != nil {
		http.Error(w, "Error querying historical stats", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var stat HistoricalPageStat
		err := rows.Scan(&stat.PageURL, &stat.PageViews, &stat.Bounces)
		if err != nil {
			http.Error(w, "Error reading historical stats", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		// Fetch and append GeoStats here for each PageURL within the historical context
		geoQuery := fmt.Sprintf(`
            SELECT GeoCity, GeoState, GeoCountry, SUM(PageViews) AS Views
            FROM aggregated_geolocation
            WHERE Datestamp >= DATE('now', '-%d days') AND PageURL = '%s'
            GROUP BY GeoCity, GeoState, GeoCountry;
        `, days, stat.PageURL)

		geoRows, err := db.Query(geoQuery)
		if err != nil {
			http.Error(w, "Error querying geolocation stats", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		for geoRows.Next() {
			var geoStat GeoStat
			err := geoRows.Scan(&geoStat.City, &geoStat.State, &geoStat.Country, &geoStat.Views)
			if err != nil {
				http.Error(w, "Error reading geolocation stats", http.StatusInternalServerError)
				log.Println(err)
				return
			}
			stat.GeolocationStats = append(stat.GeolocationStats, geoStat)
		}
		geoRows.Close()

		statsByPageURL = append(statsByPageURL, stat)
	}

	historicalStats.StatsByPageURL = statsByPageURL
	json.NewEncoder(w).Encode(historicalStats)
}
