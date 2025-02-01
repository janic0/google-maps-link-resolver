package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Ok    bool  `json:"ok"`
	Error Error `json:"error"`
}

type Result struct {
	PlaceType   string  `json:"place_type"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lon"`
	ZoomParam   string  `json:"zoom"`
	SearchQuery string  `json:"query"`
}

type SuccessResponse struct {
	Ok   bool   `json:"ok"`
	Data Result `json:"result"`
}

func main() {

	app := http.NewServeMux()

	app.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		now := time.Now()

		headers := w.Header()
		headers.Set("content-type", "application/json")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 1, Message: "failed to read request body"}})
			w.Write(encodedError)
			fmt.Println(string(encodedError))

			return
		}

		parsedUrl, err := url.Parse(string(body))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 2, Message: "provided url is not valid"}})
			fmt.Println(string(encodedError))
			w.Write(encodedError)
			return
		}

		parsedGoogleLocationUrl := parsedUrl

		if parsedUrl.Host == "maps.app.goo.gl" {
			req, err := http.NewRequest("GET", parsedUrl.String(), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 4, Message: "failed to initialize request"}})
				w.Write(encodedError)
				fmt.Println(string(encodedError))
				return
			}

			transport := http.Transport{}
			transport.TLSClientConfig = &tls.Config{}
			resp, err := transport.RoundTrip(req)

			if err != nil {
				w.WriteHeader(http.StatusBadGateway)
				fmt.Println(err.Error())
				encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 5, Message: "request to google maps failed"}})
				fmt.Println(string(encodedError))
				w.Write(encodedError)
				return
			}

			locationHeader := resp.Header.Get("Location")
			parsedGoogleLocationUrl, err = url.Parse(locationHeader)
			if len(locationHeader) == 0 || err != nil {
				w.WriteHeader(http.StatusBadGateway)
				encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 6, Message: "google maps provided an invalid response"}})
				fmt.Println(string(encodedError))
				w.Write(encodedError)
				return
			}
		} else if parsedUrl.Host != "www.google.com" {

			w.WriteHeader(http.StatusBadRequest)
			encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 3, Message: "url is not a google maps url"}})
			fmt.Println(string(encodedError))
			w.Write(encodedError)
			return

		}

		urlSections := strings.Split(parsedGoogleLocationUrl.Path, "/")

		// url is /maps/search
		if len(urlSections) == 4 && urlSections[2] == "search" {
			coordinateString := urlSections[3]
			coordinateSections := strings.Split(coordinateString, ",")
			if len(coordinateSections) < 2 {
				w.WriteHeader(http.StatusBadGateway)
				encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 10, Message: "google maps provided an invalid response"}})
				fmt.Println(string(encodedError))
				w.Write(encodedError)
				return
			}

			latitude, latitudeError := strconv.ParseFloat(coordinateSections[0], 64)
			longitude, longitudeError := strconv.ParseFloat(coordinateSections[1][1:], 64)

			if latitudeError != nil || longitudeError != nil {
				w.WriteHeader(http.StatusBadGateway)
				encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 11, Message: "google maps provided invalid coordinates"}})
				fmt.Println(string(encodedError))
				w.Write(encodedError)
				return
			}

			encodedResultResponse, _ := json.Marshal(SuccessResponse{Ok: true, Data: Result{Latitude: latitude, Longitude: longitude, PlaceType: "search"}})
			w.Write([]byte(encodedResultResponse))

			return
		}

		if len(urlSections) < 5 {
			w.WriteHeader(http.StatusBadGateway)
			encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 7, Message: "google maps provided an invalid response"}})
			fmt.Println(string(encodedError))
			w.Write(encodedError)
			return
		}

		coordinateBit := urlSections[4]
		placeType := "place"
		// edge case for direct urls like /maps/@xx,xx,xx
		if urlSections[2][0] == '@' {
			coordinateBit = urlSections[2]
			placeType = "place-direct"
		}

		coordinateSections := strings.Split(coordinateBit, ",")

		if len(coordinateSections) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 8, Message: "google maps provided an invalid response"}})
			fmt.Println(string(encodedError))
			w.Write(encodedError)
			return
		}

		latitude, latitudeError := strconv.ParseFloat(coordinateSections[0][1:], 64)
		longitude, longitudeError := strconv.ParseFloat(coordinateSections[1], 64)

		if latitudeError != nil || longitudeError != nil {
			w.WriteHeader(http.StatusBadGateway)
			encodedError, _ := json.Marshal(ErrorResponse{Ok: false, Error: Error{Code: 9, Message: "google maps provided invalid coordinates"}})
			fmt.Println(string(encodedError))
			w.Write(encodedError)
			return
		}

		unescapedSearchQuery, err := url.QueryUnescape(urlSections[3])
		if err != nil {
			unescapedSearchQuery = ""
		}

		encodedResultResponse, _ := json.Marshal(SuccessResponse{Ok: true, Data: Result{Latitude: latitude, Longitude: longitude, ZoomParam: coordinateSections[2], SearchQuery: unescapedSearchQuery, PlaceType: placeType}})
		w.Write([]byte(encodedResultResponse))

		fmt.Println("query handled | "+r.URL.Path, " | ", time.Since(now).Milliseconds(), "ms")

	})

	log.Fatal(http.ListenAndServe("0.0.0.0:8000", app))

}
