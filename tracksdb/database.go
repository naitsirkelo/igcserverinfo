package tracksdb

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// -----------------
var Global_db TracksStorage

// TracksMongoDB stores the details of the DB connection.
type TracksMongoDB struct {
	DatabaseURL            string
	DatabaseName           string
	TracksCollectionName string
}

/*
Init initializes the mongo storage.
*/
func (db *TracksMongoDB) Init() {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	index := mgo.Index{
		Key:        []string{"date"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = session.DB(db.DatabaseName).C(db.TrackssCollectionName).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

/*
Add adds new tracks to the storage.
*/
func (db *TracksMongoDB) Add(t Track) error {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.TrackssCollectionName).Insert(t)
	if err != nil {
		fmt.Printf("error in Insert(): %v", err.Error())
		return err
	}

	return nil
}

/*
Count returns the current count of the tracks in in-memory storage.
*/
func (db *TracksMongoDB) Count() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// handle to "db"
	count, err := session.DB(db.DatabaseName).C(db.TrackssCollectionName).Count()
	if err != nil {
		fmt.Printf("error in Count(): %v", err.Error())
		return -1
	}

	return count
}

/*
Get returns a student with a given ID or empty track struct.
*/
func (db *TracksMongoDB) Get(keyID string) (Track, bool) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	track := Track{}

	err = session.DB(db.DatabaseName).C(db.TrackssCollectionName).Find(bson.M{"date": H_date}).One(&track)
	if err != nil {
		return track, false
	}

	return track, true
}

/*
GetAll returns a slice with all the tracks.
*/
func (db *TracksMongoDB) GetAll() []Track {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var all []Track

	err = session.DB(db.DatabaseName).C(db.TrackssCollectionName).Find(bson.M{}).All(&all)
	if err != nil {
		return []Track{}
	}

	return all
}
