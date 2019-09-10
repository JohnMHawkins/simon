package wtmcache

import (
	"testing"
	"time"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"	
)

/////////////////////////
// Test Doc structure

type DocTest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Foo  string `json:"foo"`
}

func (d *DocTest) GetKey() (string, error) {
	return d.ID, nil
}

func NewDocFunc() *DocTest {
	doc := DocTest{}
	return &doc
}

//
/////////////////////////

/////////////////////////
// Test globals and setup

var session *mgo.Session = nil

// make a global connection to the db
func ensureSession() (*mgo.Session, error) {
	s, e := mgo.Dial("127.0.0.1:27017")
	session = s
	return s, e
}

func closeSession() {
	session.Close()
}

// get the test db and drop wtcache
func dropTestCollections() {
	db := session.DB("test")
	db.C("wtmcache").DropCollection()
}

// create the wtcache collection with an index
func createTestCollection() {
	index := mgo.Index{Key: []string{"id"}, Unique: true}
	db := session.DB("test")
	db.C("wtmcache").EnsureIndex(index)
//	db.C("wtmcachepersist").EnsureIndex(index)


}

func populateTestCollection() {
	db := session.DB("test")
	db.C("wtmcache").DropCollection()
	doc := DocTest{ID: "aaa", Name: "Able", Foo: "xxx"}
	db.C("wtmcache").Upsert(bson.M{"id": doc.ID}, doc)
	doc = DocTest{ID: "bbb", Name: "Baker", Foo: "xxx"}
	db.C("wtmcache").Upsert(bson.M{"id": doc.ID}, doc)
	doc = DocTest{ID: "ccc", Name: "Charlie", Foo: "yyy"}
	db.C("wtmcache").Upsert(bson.M{"id": doc.ID}, doc)

}

/////////////////////////////////
// Tests


// Fetch DB by session
func TestGetDbCache(t *testing.T) {
	_, err := ensureSession()
	if err != nil {
		t.Fatalf("TestGetDbCache failed: %s", err)
	}
	dropTestCollections()
	createTestCollection()	
	cDb := NewDb(session, "test")
	if cDb == nil  {
		t.Fatalf("TestGetDbCache, cDB is nil")
	}
	closeSession()
}


// test read existing collection
func TestReadExistingCollection(t *testing.T) {
	_, _ = ensureSession()
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcachepersist", "id", 5*time.Minute, 5*time.Minute)
	if cColl == nil {
		t.Fatalf("TestReadExistingCollection failed with nil cColl")
	}
	var fetchDoc = DocTest{}
	var ok bool
	r, err2 := cColl.Read("123", &fetchDoc)
	if err2 != nil {
		t.Fatalf("TestReadExistingCollection read error %s", err2)
	}
	fetchDoc, ok = r.(DocTest)
	if ok && fetchDoc.Name != "Fred" {
		t.Fatalf("TestReadExistingCollection expected name = Fred, returned %s", fetchDoc.Name)
	}	
	closeSession()	
}

// test write fast
func TestWriteFastToCollection(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
	if cColl == nil {
		t.Fatalf("TestWriteToCollection failed with nil cColl")
	}

	doc := DocTest{ID: "xyz", Name: "Joe", Foo: "bar"}
	err := cColl.WriteFast(doc.ID, doc)
	if err != nil {
		t.Fatalf("TestWriteToCollection failed to write with err= %s", err)
	}

	// now re-read it
	var fetchDoc = DocTest{}
	var ok bool
	r, err2 := cColl.Read(doc.ID, &fetchDoc)
	if err2 != nil {
		t.Fatalf("TestWriteToCollection read back error %s", err2)
	}
	fetchDoc, ok = r.(DocTest) 
	if ok && fetchDoc.Name != "Joe" {
		t.Fatalf("TestWriteToCollection expected name = Joe, returned %s", fetchDoc.Name)
		t.Fatalf("%w", fetchDoc)
	}	
	closeSession()	
}


// test write
func TestWriteToCollection(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
	if cColl == nil {
		t.Fatalf("TestWriteToCollection failed with nil cColl")
	}
	doc := DocTest{ID: "xyz", Name: "Joe", Foo: "bar"}
	err := cColl.Write(doc)
	if err != nil {
		t.Fatalf("TestWriteToCollection failed to write with err= %s", err)
	}
	// now re-read it
	var fetchDoc = DocTest{}
	var ok bool
	r, err2 := cColl.Read("xyz", &fetchDoc)
	if err2 != nil {
		t.Fatalf("TestWriteToCollection read back error %s", err2)
	}
	fetchDoc, ok = r.(DocTest) 
	if ok && fetchDoc.Name != "Joe" {
		t.Fatalf("TestWriteToCollection expected name = Joe, returned %s", fetchDoc.Name)
		t.Fatalf("%w", fetchDoc)
	}	
	closeSession()	
}


// ensure we're reading from cache by writing then out of band deleting db entry underneath cache
func TestEnsureCacheRead(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
	if cColl == nil {
		t.Fatalf("TestEnsureCacheRead failed with nil cColl")
	}
	doc := DocTest{ID: "xyz", Name: "Joe", Foo: "bar"}
	err := cColl.Write(doc)
	if err != nil {
		t.Fatalf("TestEnsureCacheRead failed to write with err= %s", err)
	}
	// delete the db underneath
	dropTestCollections()
	// now re-read it
	var fetchDoc = DocTest{}
	var ok bool
	r, err2 := cColl.Read("xyz", &fetchDoc)
	if err2 != nil {
		t.Fatalf("TestEnsureCacheRead read back error %s", err2)
	}
	fetchDoc, ok = r.(DocTest) 
	if ok && fetchDoc.Name != "Joe" {
		t.Fatalf("TestEnsureCacheRead expected name = Joe, returned %s", fetchDoc.Name)
		t.Fatalf("%w", fetchDoc)
	}	
	closeSession()	
}


// test that read with an uninitialized cache reads into the cache from the db
func TestInitialeCacheOnRead(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
	if cColl == nil {
		t.Fatalf("TestInitialeCacheOnRead failed with nil cColl")
	}
	doc := DocTest{ID: "xyz", Name: "Joe", Foo: "bar"}
	err := cColl.Write(doc)
	if err != nil {
		t.Fatalf("TestInitialeCacheOnRead failed to write with err= %s", err)
	}
	// clear the cache
	cColl.C.Flush()
	// now re-read it
	var fetchDoc = DocTest{}
	var ok bool
	r, err2 := cColl.Read("xyz", &fetchDoc)
	if err2 != nil {
		t.Fatalf("TestInitialeCacheOnRead read back error %s", err2)
	}
	fetchDoc, ok = r.(DocTest) 
	if ok && fetchDoc.Name != "Joe" {
		t.Fatalf("TestInitialeCacheOnRead expected name = Joe, returned %s", fetchDoc.Name)
		t.Fatalf("%w", fetchDoc)
	}	
	// ensure the cache exists
	r, present := cColl.C.Get("xyz")
	if (!present || r == nil) {
		t.Fatalf("TestInitialeCacheOnRead didn't initialized cache")
	}
	closeSession()	
}


// test delete entry
func TestDeleteEntry(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	populateTestCollection()
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)
	// make sure we can read Baker
	var fetchDoc = new(DocTest)
	r, err := cColl.Read("bbb", fetchDoc)
	fetchDoc, ok := r.(*DocTest)
	if !ok  {
		t.Fatalf("TestDeleteEntry not set up correctly, Baker entry doesn't exist")
	}
	// now delete baker
	err2 := cColl.Delete("bbb")
	if ( err2 != nil) {
		t.Fatalf("TestDeleteEntry Delete call failed with error: %s", err2)
	}
	// now refetch to make sure it's gone
	r, err = cColl.Read("bbb", &fetchDoc)
	if err == nil {
		t.Fatalf("TestDeleteEntry failed, still able to read the entry")
	}
	// ensure it's gone from the db and the cache as well ?

	closeSession()	

}


// test read ignore cache
func TestReadIgnoreCache(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	populateTestCollection()
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)

	// db has key = bbb as "Baker".  add a bogus value to the cache that is "beta"
	fakedoc := DocTest{ID: "bbb", Name: "Beta", Foo: "err"}
	cColl.C.Set("bbb", fakedoc, 1*time.Hour)

	// now fetch ignoring cache
	var fetchDoc = new(DocTest)
	r, err := cColl.ReadNoCache("bbb", fetchDoc)
	if err != nil {
		t.Fatalf("TestReadIgnoreCache failed to read doc, err: %s", err)
	}
	fetchDoc, _ = r.(*DocTest)
	if fetchDoc.Name != "Baker" {
		t.Fatalf("TestReadIgnoreCache still read cache")
	}
	closeSession()	
}




// test read entire collection
func TestReadEntireCollection(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	populateTestCollection()
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)

	// fetch then entire collection
	var results []DocTest
	var query interface{}
	err := cColl.Query(query, &results)
	if err != nil {
		t.Fatalf("TestReadIgnoreCache query failed with err: %s", err)
	}
	// check to see if all three made it
	if ( len(results) != 3 ) {
		t.Fatalf("TestReadEntireCollection read collection not equal to 3: size is %d", len(results))
	}
	for i, v := range results {
		switch i {
		case 0:
			if v.ID != "aaa" {
				t.Fatalf("TestReadEntireCollection first value not aaa, is %s", v.ID)
			}
		case 1:
			if v.ID != "bbb" {
				t.Fatalf("TestReadEntireCollection first value not bbb, is %s", v.ID)
			}
		case 2:
			if v.ID != "ccc" {
				t.Fatalf("TestReadEntireCollection first value not ccc, is %s", v.ID)
			}

		}
	}


	closeSession()	
}



// test query collection
func TestQueryCollection(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	populateTestCollection()
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)

	// fetch then entire collection
	var results []DocTest
	var query = bson.M{"foo" : "xxx"}
	err := cColl.Query(query, &results)
	if err != nil {
		t.Fatalf("TestQueryCollection query failed with err: %s", err)
	}
	// check to see if two made it
	if ( len(results) != 2 ) {
		t.Fatalf("TestQueryCollection read collection not equal to 2: size is %d", len(results))
	}
	for i, v := range results {
		switch i {
		case 0:
			if v.ID != "aaa" {
				t.Fatalf("TestQueryCollection first value not aaa, is %s", v.ID)
			}
		case 1:
			if v.ID != "bbb" {
				t.Fatalf("TestQueryCollection first value not bbb, is %s", v.ID)
			}
		}
	}


	closeSession()	
}



// test read sorted collection
func TestQuerySortedCollection(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	populateTestCollection()
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)

	// fetch then entire collection
	var results []DocTest
	var query = bson.M{"foo" : "xxx"}
	sortBy := "-name"
	err := cColl.QueryAndSort(query, &results, sortBy)
	if err != nil {
		t.Fatalf("TestQueryCollection query failed with err: %s", err)
	}
	// check to see if two made it
	if ( len(results) != 2 ) {
		t.Fatalf("TestQueryCollection read collection not equal to 2: size is %d", len(results))
	}
	for i, v := range results {
		switch i {
		case 0:
			if v.ID != "bbb" {
				t.Fatalf("TestQueryCollection first value not bbb, is %s", v.ID)
			}
		case 1:
			if v.ID != "aaa" {
				t.Fatalf("TestQueryCollection first value not aaa, is %s", v.ID)
			}
		}
	}


	closeSession()	
}


// test read raw query
func TestRawQueryCollection(t *testing.T) {
	_, _ = ensureSession()
	dropTestCollections()
	createTestCollection()	
	populateTestCollection()
	cDb :=  NewDb(session, "test")
	cColl := cDb.NewCollection("wtmcache", "id", 5*time.Minute, 5*time.Minute)

	// fetch then entire collection
	var query = bson.M{"foo" : "xxx"}
	results := cColl.RawQuery(query)
	if results == nil {
		t.Fatalf("TestRawQueryCollection query returned nil query")
	}
	count, err := results.Count()
	if (err != nil ) {
		t.Fatalf("TestRawQueryCollection, result couldn't get a count, err = %s", err)
	}
	if (count != 2 ) {
		t.Fatalf("TestRawQueryCollection didn't return a count of 2, instead returned %d", count)
	}

	closeSession()
}


// TODO 
// test read paginated collection



// test read sorted/paginated collection

