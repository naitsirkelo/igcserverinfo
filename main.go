package main

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"strconv"
	"bytes"
	"time"
	"log"
	"os"
	"io"
	"github.com/marni/goigc"				// Main library for working on IGC files
	"github.com/p3lim/iso8601"			// For formatting time into ISO 8601
	"gopkg.in/mgo.v2"								// Connecting to MongoDB
	"github.com/gorilla/mux"				// Creating router
)


var TrackUrl map[int]string		// Declare map for storing URLs and URL key (ID)
var Timestamps map[int]string // - For storing URL ID and timestamp as string
var Webhooks map[int]string		//
var Ids []int									// Declare slice for storing IDs
var startTime time.Time				// Variable for calculating uptime
var t_latest 	string		// Globally latest
var t_start 	string		// First timestamp for added track
var t_stop 		string		// Latest timestamp in list of tracks

var mongoDBurl string
var webhookUrl string

var session *mgo.Session


type MongoDB struct {
	DatabaseURL    string
	DatabaseName   string
	Collection		 string
}


var mongoTracks = MongoDB{DatabaseURL: MONGODBURL, DatabaseName: "igc", Collection: "tracks"}
var mongoTimestamps = MongoDB{DatabaseURL: MONGODBURL, DatabaseName: "igc", Collection: "timestamps"}
var db *mgo.Database

const (
	MAXTRACKS = 5
	MONGODBURL = "mongodb://oleklar:Passord1@ds139992.mlab.com:39992/igc"
	WEBHOOKURL = "https://hooks.slack.com/services/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"
)


type Track struct {
	H_date string					`json:"date"`
	Pilot string					`json:"pilot"`
	Glider string					`json:"glider"`
	Glider_id string 			`json:"glider_id"`
	Track_length float64	`json:"track_length"`
	Track_url	string			`json:"track_src_url"`
}


type Timestamp struct {
	T_latest		string	`json:"t_latest"`
	T_start			string	`json:"t_start"`	// Første timestampet som algres
	T_stop			string	`json:"t_stop"`
	Tracks			[]int		`json:"tracks"`
	Processing	int			`json:"processing"`
}


type MetaData struct {		// Encoding meta information of server
	Uptime 	string	`json:"uptime"`
	Info 		string	`json:"info"`
	Version string	`json:"version"`
}


type TrackId struct {			// Encoding URL ID
	Id int	`json:"id"`
}


type TimestampDB struct {
	timestamps map[string]Timestamp
}

type TrackDB struct {
	tracks map[string]Track
}

type WebhookUrl struct {
	Type 	string		`json:"type"`
}

type MinTriggerValue struct {
	Type 	int				`json:"value"`
}

type WebhookInfo struct {
	WebhookUrl 				string 	`json:"webhookURL"`
	MinTriggerValue 	int			`json:"minTriggerValue"`
}


func handleTrackPlus(w http.ResponseWriter, r *http.Request) {

	if (r.Method == http.MethodGet) {	// Check if GET was called

		parts := strings.Split(r.URL.Path, "/")	// Storing URL parts in a new array
		l := len(parts)				// Number of parts
		t := len(TrackUrl)		// Number of already stored IDs
		idString := parts[4]	// Stores ID from 5th element in a string

		id, err := strconv.Atoi(idString)	// Converts to int
		if(err != nil){
				http.Error(w, "Converting Failed. Request Timeout.", 408)
		}

		url := TrackUrl[id]	// Gets the correct URL based on input id

		track, err2 := igc.ParseLocation(url)	// Parses information from URL location
		if(err2 != nil){
				http.Error(w, "Parsing Failed. Request Timeout.", 408)
		}																			// Calculates track length
		trackLen := track.Points[0].Distance(track.Points[len(track.Points)-1])

 		if (id > 0 && id <= t) {
		// WITH FIELD
			if (l == 6) {					// If the call consists of 6 parts

				field := parts[5]				// Store the field input

				if (field == "pilot") {	// Check which variable 'field' is equal to
					fmt.Fprintln(w, track.Pilot)

				} else if (field == "glider") {
					fmt.Fprintln(w, track.GliderType)

				} else if (field == "glider_id") {
					fmt.Fprintln(w, track.GliderID)

				} else if (field == "track_length") {
					fmt.Fprintln(w, trackLen)

				} else if (field == "H_date") {
					fmt.Fprintln(w, track.Date.String())

				} else if (field == "track_src_url") {
					fmt.Fprintln(w, url)

				} else {	// If neither of the correct variables are called: 400
					http.Error(w, "No valid Field value.", 400)	// Bad request
				}

			// ONLY ID, EMPTY FIELD
			} else if (l == 5) {	// 5 parts and id input is valid (1-len)

									// Creating temporary struct to hold variables
				temp := Track{track.Date.String(), track.Pilot, track.GliderType,
													track.GliderID, trackLen, url}
									// Encodes temporary struct and shows information on screen
				http.Header.Add(w.Header(), "Content-type", "application/json")
				err3 := json.NewEncoder(w).Encode(temp)
				if(err3 != nil){
				  	http.Error(w, "Encoding Failed. Request Timeout.", 408)
				}
			// ELSE: Invalid ID entered
			} else {
				http.Error(w, "No valid ID value.", 400)	// Bad request
			}

		} else {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}
}


func handleTrack(w http.ResponseWriter, r *http.Request) {

		if (r.Method == http.MethodGet) {		// Check if GET was called

			updateTrackIds()
			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(Ids)		// Encode ID int array
			if(err != nil){
			  	http.Error(w, "Encoding Failed. Request Timeout.", 408)
			}

		} else if (r.Method == http.MethodPost) {	// Check if POST was called

			var temp map[string]interface{}	// Interface is unknown type / C++ auto

			if (len(TrackUrl) == 0) {
				t_start = iso8601.Format(time.Now())	// Store first timestamp
			}
			t_latest = iso8601.Format(time.Now())		// Update latest timestamp

			err2 := json.NewDecoder(r.Body).Decode(&temp)					// Decode posted url
			if (err2 != nil) {
				http.Error(w, "Decoding Failed. Request Timeout.", 408)
			}
			if (err2 == io.EOF) {		// Empty body error
				http.Error(w, "Empty Body For Post Request.", 400)	// Bad request
			}
															// Internal identification: Int up from 1
			tempLen := len(TrackUrl) + 1
															// Places url in map spot nr 1 and up
			TrackUrl[tempLen] = temp["url"].(string)

			idStruct := TrackId{Id: tempLen}
															// Define header for correct output
			http.Header.Add(w.Header(), "Content-type", "application/json")
			err3 := json.NewEncoder(w).Encode(idStruct)
			if (err3 != nil) {
				http.Error(w, "Encoding Failed. Request Timeout.", 408)
			}
		}
}


func handleApi(w http.ResponseWriter, r *http.Request) {
	if (r.Method == http.MethodGet) {	// Check if GET was called
									// Using included library, format time Since into ISO 8601
		t := iso8601.Format(time.Since(startTime))

									// Prepare struct for encoding
		temp := MetaData{t, "Service for Paragliding tracks.", "v1"}

									// Define header for correct output
		http.Header.Add(w.Header(), "Content-type", "application/json")
		err := json.NewEncoder(w).Encode(temp)
		if(err != nil){
				http.Error(w, "Encoding Failed. Request Timeout.", 408)
		}
	}
}


func handleTicker(w http.ResponseWriter, r *http.Request) {
	if (r.Method == http.MethodGet) {					// Check if GET was called

		parts := strings.Split(r.URL.Path, "/")	// Storing URL parts in a new array
		l := len(parts)													// Number of parts

		if (l == 5) {									// If the call consists of 6 parts
																	// Stores ID from 6th element in a string
			webhookId := parts[5]
			if (parts[4] == "latest") {
				// GET /api/ticker/latest
				fmt.Fprintln(w, "\nNumber of Tracks in DB: ", t_latest)

			} else {
				// GET /api/ticker/<timestamp>
				timestamp := parts[4]


			}
		} else {

			updateTrackIds()
			processing := 0

			temp := Timestamp{t_latest, t_start,
											  t_stop, Ids, processing}

			http.Header.Add(w.Header(), "Content-type", "application/json")

			err := json.NewEncoder(w).Encode(temp)
			if(err != nil){
					http.Error(w, "Encoding Failed. Request Timeout.", 408)
			}
		}
	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}


func handleWebhook(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/")	// Storing URL parts in a new array
	l := len(parts)					// Number of parts

	if (l == 6) {						// If the call consists of 6 parts
		webhookId := parts[5]	// Stores ID from 6th element in a string
	}

	if (r.Method == http.MethodGet) {	// Check if GET was called
		temp := WebhookInfo{}

		temp.WebhookUrl = WEBHOOKURL

		raw, _ := json.Marshal(temp)

		resp, err := http.Post(MONGODBURL, "application/json", bytes.NewBuffer(raw))
		if err != nil {
			fmt.Println(err)
			fmt.Println(ioutil.ReadAll(resp.Body))
		}

	} else if (r.Method == http.MethodPost) {

		temp := WebhookInfo{}

		temp.Webhookurl = WEBHOOKURL
		temp.Mintriggervalue = 1

		raw, _ := json.Marshal(temp)
		resp, err := http.Post(WEBHOOKURL, "application/json", bytes.NewBuffer(raw))
		if err != nil {
			fmt.Println(err)
			fmt.Println(ioutil.ReadAll(resp.Body))
	}

	} else if (r.Method == http.MethodDelete) {




	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}


func handleAdmin(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/")	// Storing URL parts in a new array
	l := len(parts)													// Number of parts

	if (r.Method == http.MethodGet && parts[4] == "tracks_count") {
																// Check if GET was used and correct path called
		n := db.Count()
		fmt.Fprintln(w, "Number of Tracks in DB: ", n)

	} else if (r.Method == http.MethodDelete && parts[4] == "tracks") {
																// Check if DELETE was used and correct path called
		trackSlice := db.GetAll()
		trackSlice = trackSlice[:0]	// Empties the track slice

		fmt.Fprintln(w, "All Tracks deleted from DB.")

	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

				// Handle every call beginning with anything else than igcinfo/api
func handleInvalid(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func redirectApi(w http.ResponseWriter, r *http.Request) {
		// http.Redirect(w, r, "http://www.golang.org", 301)
		http.HandleFunc("/paragliding/api", handleApi)
}


func tickerFunc() {
	ticker := time.NewTicker(500 * time.Millisecond)
    go func() {
        for t := range ticker.C {
            fmt.Println("Tick at", t)
        }
    }()

    time.Sleep(1600 * time.Millisecond)
    ticker.Stop()
    fmt.Println("Ticker stopped")
}


func updateTrackIds() {
	Ids = Ids[:0]											// Reseting slice before appending keys
	for key, url := range TrackUrl {	// Append each key in TrackUrl to the slice 'a'
		Ids = append(Ids, key)
		fmt.Println("\n", url)					// To avoid console error of URL not used.
	}
}


func getPort() string {
	 	var port = os.Getenv("PORT")
 				// Port sets to :8080 as a default
 		if (port == "") {
 			port = "8080"
			fmt.Println("No PORT variable detected, defaulting to " + port)
 		}
 		return (":" + port)
}


func init() {

	session, err = mgo.Dial(db.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	c := session.DB(database).C(collection)
	err2 := c.Find(query).One(&result)
	if err != nil {
		log.Fatal(err2)
	}
	session.SetSafe(&mgo.Safe{})

	TrackUrl = make(map[int]string) 	// Initializing map arrays
	Webhooks = make(map[int]string)		//

	Ids = make([]int, len(TrackUrl))	// Initialize empty ID int slice
	startTime = time.Now()						// Initializes timer

	slackUrl = "https://oleklar.slack.com/messages/CDJ5VV12T/"
}


func main() {

		if len(mongoDBurl) == 0 {
			log.Fatal("PARAGLIDING_MONGO environment variable is not set (put mongodb url in here)")
		}

		trackdb.Global_db = &trackdb.TracksMongoDB{
			mongoDBurl,
			"tracksDB",
			"tracks",
		}

		port := getPort()

		init()
		r := mux.NewRouter()

		r.HandleFunc("/", handleInvalid)
		r.HandleFunc("/paragliding/api/", handleInvalid)

		r.HandleFunc("/paragliding/", handleApi)
		r.HandleFunc("/paragliding/api", handleApi)

		r.HandleFunc("/paragliding/api/track", handleTrack)
		r.HandleFunc("/paragliding/api/track/", handleTrackPlus)

		r.HandleFunc("/paragliding/api/ticker/", handleTicker)
		r.HandleFunc("/paragliding/api/webhook/new_track", handleWebhook)

		r.HandleFunc("/paragliding/admin/api/", handleAdmin)

		if err := http.ListenAndServe(port, r); err != nil {
			log.Fatal(err)
		}
}
