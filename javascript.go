package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func serveJavaScript(domain, port string, w http.ResponseWriter) {
	collectURL := fmt.Sprintf("//%s:%s/collect", domain, port)
	if port == "80" || port == "443" {
		// For standard ports, omit the port number in the URL
		collectURL = fmt.Sprintf("//%s/collect", domain)
	}

	js := fmt.Sprintf(`
(function() {
    // Capture page load event
    document.addEventListener("DOMContentLoaded", function() {
        var data = {
            url: window.location.href,
            referrer: document.referrer,
            userAgent: navigator.userAgent,
            // Extend with additional data points as needed
        };

        // Send the data to the server
        fetch("%s", {
            method: "POST",
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        })
        .then(response => console.log('Analytics data sent'))
        .catch(error => console.log('Error sending analytics data', error));
    });
})();
`, collectURL)

	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(js))
}

func collectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// Only accept POST requests
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Add additional data processing here if necessary
	// For example, extracting geolocation data from the request

	// Log the collected data
	logRequest(data)

	// Send a response back to the client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Data collected"))
}
