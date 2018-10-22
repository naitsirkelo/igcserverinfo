package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"strings"
	"strconv"
	"time"
	"log"
	"os"
	"io"
	"github.com/naitsirkelo/igcserverinfo"
	"github.com/marni/goigc"			// Main library for working on IGC files
	"github.com/p3lim/iso8601"		// For formatting time into ISO 8601
	// "github.com/BurntSushi/toml"	// For accessing config file
	// mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	// "github.com/gorilla/mux"
	// "gopkg.in/mgo.v2-unstable/internal/scram"
	// "github.com/Delostik/SimpleWebHook"
)


var TrackUrl map[int]string		// Declare map for storing URLs and URL key (ID)
var Timestamps map[int]string // - For storing URL ID and timestamp as string
var Webhooks map[int]string		//
var Ids []int									// Declare slice for storing IDs
var startTime time.Time				// Variable for calculating uptime
var t_latest time.Time
var t_start time.Time
var t_stop time.Time
var notStarted time.Time

var config = Config{}
var dao = TracksDAO{}
var db *mgo.Database

const (
	COLLECTION_1 = "tracks"
	COLLECTION_2 = "timestamps"
	MAXTRACKS = 5
)


// const (
// 	TICKERINTERVAL = time.Second * 24
// )


type Config struct {
	Server   string
	Database string
}


type Track struct {
	Id bson.Object 				`bson:"_id,omitempty"`
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
				temp := TrackInfo{track.Date.String(), track.Pilot, track.GliderType,
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

			err2 := json.NewDecoder(r.Body).Decode(&temp)					// Decode posted url
			if (err2 != nil) {
				http.Error(w, "Decoding Failed. Request Timeout.", 408)
			}
			if (err2 == io.EOF) {	// Empty body error
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

		if (l == 5) {						// If the call consists of 6 parts
			// webhookId := parts[5]	// Stores ID from 6th element in a string

			if (parts[4] == "latest") {
				// GET /api/ticker/latest



			} else {
				// GET /api/ticker/<timestamp>
				// timestamp := parts[4]



			}
		} else {
			// GET /api/ticker/



		}
	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}


func handleWebhook(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/")	// Storing URL parts in a new array
	l := len(parts)					// Number of parts

	if (l == 6) {						// If the call consists of 6 parts
		// webhookId := parts[5]	// Stores ID from 6th element in a string
	}

	if (r.Method == http.MethodGet) {	// Check if GET was called




	} else if (r.Method == http.MethodPost) {
		var temp WebhookStruct

		err := json.NewDecoder(r.Body).Decode(&temp)
		if err != nil {
			http.Error(w, "Decoding Failed. Request Timeout.", 400)	// Bad Request
		}

	} else if (r.Method == http.MethodDelete) {




	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}


func handleAdmin(w http.ResponseWriter, r *http.Request) {

	if (r.Method == http.MethodGet) {	// Check if GET was called



	} else if (r.Method == http.MethodDelete) {



	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

				// Handle every call beginning with anything else than igcinfo/api
func handleInvalid(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

				// Find the correct Port to run app on Heroku
func getPort() string {
	 	var port = os.Getenv("PORT")
 				// Port sets to :8080 as a default
 		if (port == "") {
 			port = "8080"
			fmt.Println("No PORT variable detected, defaulting to " + port)
 		}
 		return (":" + port)
}


// func (m *Tracks) Connect() {
// 	session, err := mgo.Dial(m.Server)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	db = session.DB(m.Database)
// }


func redirectApi(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://www.golang.org", 301)
		// http.HandleFunc("/paragliding/api", handleApi)
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


type MongoDB struct {
	DatabaseURL    string
	DatabaseName   string
	CollectionName string
}


type TimestampMongoDB struct {
	DatabaseURL   					string
	DatabaseName 						string
	TimestampCollectionName string
}


type TrackMongoDB struct {
	DatabaseURL   			string
	DatabaseName 				string
	TrackCollectionName string
}


type TimestampDB struct {
	timestamps map[string]Timestamp
}


type TrackDB struct {
	tracks map[string]TrackInfo
}


func (db *TrackMongoDB) Add(t TrackInfo) {
	session, err := mgo.Dial(db.DatabaseName)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Insert document
	session.DB(db.DatabaseName).C(db.TrackCollectionName).Insert()
	if err2 != nil {
		// Return log
		log("Error in Insert():", err.Error())
	}
}


func (db *TrackMongoDB) Count() {
	session, err := mgo.Dial(db.DatabaseName)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	count, err := session.DB(db.DatabaseName).C(db.TrackCollectionName).Count()
	if err2 != nil {
		// Return log
		log("Error in Count():", err.Error())
		return -1
	}
	return count
}


func (db *TrackMongoDB) Search(urlID int) (TrackInfo, bool){
	session, err := mgo.Dial(db.DatabaseName)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	temp_track := TrackInfo{}

	err2 = session.DB(db.DatabaseName).C(db.TrackCollectionName).Find(bson.M{"Track_url": urlID}).One(&temp_track)
	if err2 != nil {
		return temp_track, false	// Empty student and false result
	}
	return temp_track, true			// Found student and true result
}


func init() {
	TrackUrl = make(map[int]string) 	// Initializing map arrays
	Webhooks = make(map[int]string)		//

	Ids = make([]int, len(TrackUrl))	// Initialize empty ID int slice
	startTime = time.Now()						// Initializes timer
	// ticker = time.NewTicker(time.Millisecond * 500)	// Initialize 'ticker'

	config.readConfig()

	session, err := mgo.dial(db.DatabaseName)
	if err != nil {
		panic(err)
	}
	defer session.Close()



	db.timestamps = make(map[string]Timestamp)
	db.tracks = make(map[string]TrackInfo)

}


func (db *MongoDB) mongoInit() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"studentid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.CollectionName).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}



func setupDB *TrackMongoDB {
	db := TrackMongoDB{
		DatabaseURL: 	"mongodb://localhost",
		DatabaseName: "tracksDB",
		TrackCollectionName "tracks"
	}

	session, err := mgo.dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	return &db
}


func main() {
	studentdb.Global_db = &studentdb.StudentsDB{}

	// Using MongoDB based storage
/* studentdb.Global_db = &studentdb.StudentsMongoDB{
	"mongodb://localhost",
	"studentsDB",
	"students",
}*/

		trackdb.Global_db.Init()

		port := os.Getenv("PORT")

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

		http.ListenAndServe(":"+port, nil)
		// if err := http.ListenAndServe(":3000", r); err != nil {
		// 	log.Fatal(err)
		// }

}
