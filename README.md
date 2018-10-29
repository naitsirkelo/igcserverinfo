# GetIgcServerInfo
(Assignment 2 - IMT2681)

![Build Status](https://img.shields.io/badge/build-Non--Operational-yellow.svg)

The goal of this project is to make an IGC paragliding information service cloud-enabled through Heroku, utilizing MongoDB for storing URLs, timestamps and paragliding tracks.

### main.go

The file is commented all the way through in a minimalistic manner, to easiest explain the meaning behind the structure without disrupting the flow or making a clutter of the code.

### Design decisions

For modularization purposes the project functinalities are split between 3 folders: src, tracksdb and mongodb.

Storing the IGC-track Url is done with a map of integers and strings, where each URL is given an int to function as Track ID.
```
var TrackUrl map[int]string
```

The following map contains timestamps converted to strings, and an integer ID to correspond with the Track ID. This is stored by time.Now() when the track is posted.
```
var Timestamps map[int]string
```

Each webhook will receive its own ID, not corresponding with the track/timestamp ID. The string will be the URL from the POST body.
```
var Webhooks map[int]string
```

### Assumptions

We assume that each MongoDB collection are created upon launching the application, so we avoid any fatal errors that would kick in if no collections are found.

The Ticker functionality is handled in its entirety within main.go, but the following function would have been implemented with OpenStack if our solution would not be accepted.
```
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
```

## Missing features

* Deploying application and getting access through Heroku link.
* Clock trigger functionality.
* Return only **newly** added tracks when posting a webhook
* Correct modularization of project structure.


## Deployment



http://getigcserverinfo.herokuapp.com
(**Not currently running.**)

## Built With

* [goigc](github.com/marni/goigc) - Initial GitHub repository for accessing IGC information.
* [iso8601](github.com/p3lim/iso8601) - Framework for converting timestamps into the ISO8601 format.
* [mgo](gopkg.in/mgo.v2) - Utilizing MongoDB functionality.
* [mgo/bson](gopkg.in/mgo.v2/bson) - Encoding and decoding of bson.
* [Gorilla/Mux](github.com/gorilla/mux) - Creating the MongoDB session with a router.




