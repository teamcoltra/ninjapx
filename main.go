package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
	"github.com/oschwald/maxminddb-golang"
)

// Define a global variable for the MaxMind database
var maxMindDB *maxminddb.Reader

// initMaxMindDB loads the MaxMind database into memory
func initMaxMindDB(filePath string) {
	var err error
	maxMindDB, err = maxminddb.Open(filePath)
	if err != nil {
		log.Fatal("Error loading MaxMind database: ", err)
	}
}

// maxMindLookup performs a lookup of the given IP address in the MaxMind database
func maxMindLookup(ipAddress string) (map[string]interface{}, error) {
	ip := net.ParseIP(ipAddress)
	var record map[string]interface{}
	err := maxMindDB.Lookup(ip, &record)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Global database instance
var db *sql.DB

// Initialize the SQLite database
func initDB(filepath string) {
	var err error
	db, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("DB nil")
	}
	createTables(db)
	// Consider adding db.Ping() here to check for database liveliness
}

// Create tables if they do not already exist
func createTables(db *sql.DB) {
	// SQL statement for creating new tables
	sqlTable := `
    CREATE TABLE IF NOT EXISTS log(
        ID INTEGER PRIMARY KEY AUTOINCREMENT,
        Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        IPAddress TEXT,
        UserAgent TEXT,
        Referrer TEXT,
        PageURL TEXT,
        GeoCity TEXT,
		GeoState TEXT,
		GeoCountry TEXT
    );
    CREATE TABLE IF NOT EXISTS aggregated_analytics(
        ID INTEGER PRIMARY KEY AUTOINCREMENT,
        Datestamp DATE DEFAULT (date('now')),
        PageViews INTEGER,
        Bounces INTEGER,
        Referrer TEXT,
        PageURL TEXT,
        Geolocation TEXT,
        UNIQUE(Datestamp, PageURL)
    );
    CREATE TABLE IF NOT EXISTS aggregated_referrals(
        ID INTEGER PRIMARY KEY AUTOINCREMENT,
        Datestamp DATE DEFAULT (date('now')),
        ReferralCount INTEGER,
        ReferralURL TEXT,
        ReferralTLD TEXT,
        PageURL TEXT,
        Geolocation TEXT,
        UNIQUE(Datestamp, PageURL)
    );
	CREATE TABLE IF NOT EXISTS aggregated_geolocation(
        ID INTEGER PRIMARY KEY AUTOINCREMENT,
        Datestamp DATE DEFAULT (date('now')),
        GeoCity TEXT,
        GeoState TEXT,
        GeoCountry TEXT,
		PageViews INTEGER,
        PageURL TEXT, -- Marked for unique combination
        UNIQUE(Datestamp, GeoCity, GeoState, GeoCountry, PageURL)
    );
    `

	_, err := db.Exec(sqlTable)
	if err != nil {
		log.Fatal(err)
	}
}

func logRequest(data map[string]string) {
	ipAddress := data["IPAddress"]
	host, _, err := net.SplitHostPort(ipAddress)
	if err != nil {
		// If there's an error, it might be because there's no port number, so use the original IP address
		host = ipAddress
	}
	userAgent := data["UserAgent"]
	referrer := data["Referrer"]
	pageURL := data["PageURL"]

	// Perform a MaxMind lookup
	geoData, err := maxMindLookup(host)
	if err != nil {
		log.Printf("Error performing MaxMind lookup: %v", err)
		// Handle the error (e.g., proceed with logging without geolocation data)
	}

	fmt.Printf("%+v\n", geoData)

	// Initialize geolocation information variables
	var geoCity, geoState, geoCountry string

	// Safely extract geolocation information using type assertions with checks
	if city, ok := geoData["city"].(map[string]interface{}); ok {
		if names, ok := city["names"].(map[string]interface{}); ok {
			geoCity, _ = names["en"].(string)
		}
	}
	if subdivisions, ok := geoData["subdivisions"].([]interface{}); ok && len(subdivisions) > 0 {
		if subdivision, ok := subdivisions[0].(map[string]interface{}); ok {
			if names, ok := subdivision["names"].(map[string]interface{}); ok {
				geoState, _ = names["en"].(string)
			}
		}
	}
	if country, ok := geoData["country"].(map[string]interface{}); ok {
		if names, ok := country["names"].(map[string]interface{}); ok {
			geoCountry, _ = names["en"].(string)
		}
	}

	// Hash the IP address
	hasher := md5.New()
	hasher.Write([]byte(host))
	hashedIPAddress := hex.EncodeToString(hasher.Sum(nil))

	// Prepare and execute the SQL statement
	stmt, err := db.Prepare("INSERT INTO log(IPAddress, UserAgent, Referrer, PageURL, GeoCity, GeoState, GeoCountry) VALUES(?,?,?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(hashedIPAddress, userAgent, referrer, pageURL, geoCity, geoState, geoCountry)
	if err != nil {
		log.Fatal(err)
	}
}

func aggregateData() {
	// Check for any log entries from yesterday or earlier
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM log WHERE DATE(Timestamp) < DATE('now'))").Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking for log entries: %v", err)
	}
	if !exists {
		log.Println("No log entries from yesterday or earlier. Skipping aggregation.")
		return // Exit the function early
	}

	// Aggregate PageViews and Bounces per PageURL per Datestamp
	queryPageViewsAndBounces := `
    INSERT INTO aggregated_analytics (Datestamp, PageViews, Bounces, PageURL)
    SELECT
        DATE(Timestamp) AS Datestamp,
        COUNT(*) AS PageViews,
        SUM(CASE WHEN VisitCount = 1 THEN 1 ELSE 0 END) AS Bounces,
        PageURL
    FROM (
        SELECT
            PageURL,
            IPAddress,
            COUNT(*) AS VisitCount
        FROM log
        WHERE DATE(Timestamp) < DATE('now')
        GROUP BY PageURL, IPAddress
    ) AS PageVisits
    GROUP BY Datestamp, PageURL
    ON CONFLICT(Datestamp, PageURL) DO UPDATE SET
        PageViews = EXCLUDED.PageViews,
        Bounces = EXCLUDED.Bounces;
`
	err = executeQuery(db, queryPageViewsAndBounces)
	if err != nil {
		log.Fatal(err)
	}

	// Aggregate Referral data
	queryReferrals := `
		INSERT INTO aggregated_referrals (Datestamp, ReferralCount, ReferralURL, PageURL)
		SELECT
			DATE(Timestamp) AS Datestamp,
			COUNT(*) AS ReferralCount,
			Referrer AS ReferralURL,
			PageURL
		FROM log
		WHERE DATE(Timestamp) < DATE('now') AND Referrer != ''
		GROUP BY Datestamp, ReferralURL, PageURL
		ON CONFLICT(Datestamp, PageURL) DO UPDATE SET
			ReferralCount = EXCLUDED.ReferralCount;
	`
	err = executeQuery(db, queryReferrals)
	if err != nil {
		log.Fatal(err)
	}

	// Aggregate Geolocation data
	queryGeolocation := `
		INSERT INTO aggregated_geolocation (Datestamp, GeoCity, GeoState, GeoCountry, PageViews, PageURL)
		SELECT
			DATE(Timestamp) AS Datestamp,
			GeoCity,
			GeoState,
			GeoCountry,
			COUNT(*) AS PageViews,
			PageURL
		FROM log
		WHERE DATE(Timestamp) < DATE('now') 
		GROUP BY Datestamp, GeoCity, GeoState, GeoCountry, PageURL
		ON CONFLICT(Datestamp, GeoCity, GeoState, GeoCountry, PageURL) DO UPDATE SET
			PageViews = EXCLUDED.PageViews;
	`
	err = executeQuery(db, queryGeolocation)
	if err != nil {
		log.Fatal(err)
	}

	deleteQuery := "DELETE FROM log WHERE DATE(Timestamp) < DATE('now');"
	err = executeQuery(db, deleteQuery)
	if err != nil {
		log.Fatalf("Error deleting aggregated log entries: %v", err)
	}
}

func checkAndAggregate() {
	for range time.Tick(time.Hour) {
		aggregateData()
	}
}

func executeQuery(db *sql.DB, query string) error {
	_, err := db.Exec(query)
	return err
}

func main() {
	// Initialize and connect to the SQLite database
	dbPath := flag.String("db", "ninja.db", "path to the sqlite database file")
	domainName := flag.String("domain", "localhost", "The domain name of the application")
	port := flag.String("port", "8080", "The port on which the application will run")
	maxMindDBFile := flag.String("maxMindDB", "GeoLite2-City.mmdb", "MaxMind GeoLite2-City.mmdb DB File Path")

	flag.Parse()

	// Initialize the database & maxmind
	initDB(*dbPath)
	initMaxMindDB(*maxMindDBFile)

	go checkAndAggregate()

	aggregateData()

	// Setup HTTP routes
	http.HandleFunc("/pixel.gif", servePixel)
	// Handler for serving the JavaScript
	http.HandleFunc("/track.js", func(w http.ResponseWriter, r *http.Request) {
		// Pass domainName and port to serveJavaScript
		serveJavaScript(*domainName, *port, w)
	})

	// Handler for collecting analytics data
	http.HandleFunc("/collect", collectHandler)

	http.HandleFunc("/api/stats/today", serveTodayStats)
	http.HandleFunc("/api/stats/historical", serveHistoricalStats)

	address := *domainName // Start with just the domain name
	if *port != "80" {     // Append the port if it's not the default HTTP port
		address = fmt.Sprintf("%s:%s", *domainName, *port)
	}

	// Start the HTTP server
	log.Printf("Server starting on %s", address)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	select {}

}
