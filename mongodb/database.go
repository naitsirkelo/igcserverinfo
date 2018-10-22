package mongodb

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MongoDB stores the details of the DB connection.
type MongoDB struct {
	DatabaseURL    string
	DatabaseName   string
	CollectionName string
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

/*
Init initializes the mongo storage.
*/
func (db *MongoDB) Init() {
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

	err = session.DB(db.DatabaseName).C(db.CollectionName).EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

/*
Add adds new tracks to the storage.
*/
func (db *MongoDB) Add(t Track) error {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	err = session.DB(db.DatabaseName).C(db.CollectionName).Insert(t)

	if err != nil {
		fmt.Printf("error in Insert(): %v", err.Error())
		return err
	}

	return nil
}

/*
Count returns the current count of the students in in-memory storage.
*/
func (db *MongoDB) Count() int {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// handle to "db"
	count, err := session.DB(db.DatabaseName).C(db.CollectionName).Count()
	if err != nil {
		fmt.Printf("error in Count(): %v", err.Error())
		return -1
	}

	return count
}

/*
Get returns a track with a given ID or empty track struct.
*/
func (db *MongoDB) Get(keyID string) (Track, bool) {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	track := Track{}

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{"Track_url": keyID}).One(&track)
	if err != nil {
		return track, false
	}

	return track, true
}

/*
GetAll returns a slice with all the tracks.
*/
func (db *MongoDB) GetAll() []Track {
	session, err := mgo.Dial(db.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var all []Track

	err = session.DB(db.DatabaseName).C(db.CollectionName).Find(bson.M{}).All(&all)
	if err != nil {
		return []Track{}
	}

	return all
}
