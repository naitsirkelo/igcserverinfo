package tracksdb

// TracksStorage represents a unified way of accessing Student data.
type TrackStorage interface {
	Init()
	Add(t Track) error
	Count() int
	Get(key string) (Track, bool)
	GetAll() []Student
}

type Track struct {
	H_date string					`json:"date"`
	Pilot string					`json:"pilot"`
	Glider string					`json:"glider"`
	Glider_id string 			`json:"glider_id"`
	Track_length float64	`json:"track_length"`
	Track_url	string			`json:"track_src_url"`
}

/*
TracksDB is the handle to tracks in-memory storage.
*/
type TracksDB struct {
	tracks map[string]Track
}

/*
Init initializes the in-memory storage.
*/
func (db *TracksDB) Init() {
	db.tracks = make(map[string]Track)
}


/*
Count returns the current count of the tracks in in-memory storage.
*/
func (db *TracksDB) Count() int {
	return len(db.tracks)
}

/*
Get returns a track with a given ID or empty track struct.
*/
func (db *TracksDB) Get(urlID string) (Track, bool) {
	s, ok := db.tracks[urlID]
	return s, ok
}

/*
GetAll returns all the tracks as slice
*/
func (db *TracksDB) GetAll() []Track {
	all := make([]Track, 0, db.Count())
	for _, s := range db.tracks {
		all = append(all, s)
	}
	return all
}
