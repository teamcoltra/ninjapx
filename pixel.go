package main

import (
	"log"
	"net/http"
)

// Pixel data for a 1x1 transparent GIF
var pixel = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00,
	0x01, 0x00, 0x80, 0xff, 0x00, 0xff, 0xff, 0xff,
	0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

// servePixel serves the 1x1 pixel GIF and logs the request for analytics
func servePixel(w http.ResponseWriter, r *http.Request) {
	// Extract necessary data from the request
	data := make(map[string]string)
	data["IPAddress"] = r.RemoteAddr // Note: This might include a port number
	data["UserAgent"] = r.UserAgent()
	data["Referrer"] = r.Referer()
	data["PageURL"] = r.RequestURI
	// GeoCity, GeoState, GeoCountry will be filled inside logRequest

	// Log the request for analytics
	logRequest(data)

	// Set the Content-Type
	w.Header().Set("Content-Type", "image/gif")

	// Serve the 1x1 pixel GIF
	_, err := w.Write(pixel)
	if err != nil {
		// Handle the error appropriately
		log.Printf("Error serving pixel: %v", err)
	}
}
