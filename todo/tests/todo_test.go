package tests

//Introduction to testing.  Note that testing is built into go and we will be using
//it extensively in this class. Below is a starter for your testing code.  In
//addition to what is built into go, we will be using a few third party packages
//that improve the testing experience.  The first is testify.  This package brings
//asserts to the table, that is much better than directly interacting with the
//testing.T object.  Second is gofakeit.  This package provides a significant number
//of helper functions to generate random data to make testing easier.

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"drexel.edu/todo/db"
	fake "github.com/brianvoe/gofakeit/v6" //aliasing package name
	"github.com/stretchr/testify/assert"
)

// Note the default file path is relative to the test package location.  The
// project has a /tests path where you are at and a /data path where the
// database file sits.  So to get there we need to back up a directory and
// then go into the /data directory.  Thus this is why we are setting the
// default file name to "../data/todo.json"
const (
	DEFAULT_DB_FILE_NAME = "../data/todo.json"
)

var (
	DB *db.ToDo
)

// note init() is a helpful function in golang.  If it exists in a package
// such as we are doing here with the testing package, it will be called
// exactly once.  This is a great place to do setup work for your tests.
func init() {
	//Below we are setting up the gloabal DB variable that we can use in
	//all of our testing functions to make life easier
	testdb, err := db.New(DEFAULT_DB_FILE_NAME)
	if err != nil {
		fmt.Print("ERROR CREATING DB:", err)
		os.Exit(1)
	}

	DB = testdb //setup the global DB variable to support test cases

	//Now lets start with a fresh DB with the sample test data
	testdb.RestoreDB()
}

// Sample Test, will always pass, comparing the second parameter to true, which
// is hard coded as true
func TestTrue(t *testing.T) {
	assert.True(t, true, "True is true!")
}

func TestAddHardCodedItem(t *testing.T) {
	item := db.ToDoItem{
		Id:     999,
		Title:  "This is a test case item",
		IsDone: false,
	}
	t.Log("Testing Adding a Hard Coded Item: ", item)

	//finish this test, add an item to the database and then
	//check that it was added correctly by looking it back up
	//use assert.NoError() to ensure errors are not returned.
	//explore other useful asserts in the testify package, see
	//https://github.com/stretchr/testify.  Specifically look
	//at things like assert.Equal() and assert.Condition()

	//I will get you started, uncomment the lines below to add to the DB
	//and ensure no errors:
	//---------------------------------------------------------------
	err := DB.AddItem(item)
	assert.NoError(t, err, "Error adding item to DB")

	//Now finish the test case by looking up the item in the DB
	//and making sure it matches the item that you put in the DB above

	dbItem, getErr := DB.GetItem(item.Id)
	assert.NoError(t, getErr, "Error getting item from DB")
	assert.Equal(t, item, dbItem)
}

func TestAddRandomStructItem(t *testing.T) {
	//You can also use the Stuct() fake function to create a random struct
	//Not going to do anyting
	item := db.ToDoItem{}
	err := fake.Struct(&item)
	t.Log("Testing Adding a Randomly Generated Struct: ", item)

	assert.NoError(t, err, "Created fake item OK")

	addErr := DB.AddItem(item)
	assert.NoError(t, addErr, "Error adding item to DB")

	dbItem, getErr := DB.GetItem(item.Id)
	assert.NoError(t, getErr, "Error getting item from DB")

	assert.Equal(t, item, dbItem)
}

func TestAddRandomItem(t *testing.T) {
	//Lets use the fake helper to create random data for the item
	item := db.ToDoItem{
		Id:     fake.Number(100, 110),
		Title:  fake.JobTitle(),
		IsDone: fake.Bool(),
	}

	t.Log("Testing Adding an Item with Random Fields: ", item)

	addErr := DB.AddItem(item)
	assert.NoError(t, addErr, "Error adding item to DB")

	dbItem, getErr := DB.GetItem(item.Id)
	assert.NoError(t, getErr, "Error getting item from DB")

	assert.Equal(t, item, dbItem)

}

//Create additional tests to showcase the correct operation of your program
//for example getting an item, getting all items, updating items, and so on. Be
//creative here.

func TestGetAllItems(t *testing.T) {
	// First restore DB so we know exactly which items will be present
	assert.NoError(t, DB.RestoreDB(), "Error restoring DB")

	// Reload the DB from file again
	testdb, newDbErr := db.New(DEFAULT_DB_FILE_NAME)
	assert.NoError(t, newDbErr, "Error loading DB from file")

	// Get all the items
	items, err := testdb.GetAllItems()
	assert.NoError(t, err, "Error getting all items from the DB")

	// Ensure we have 4 items with ID 1-4
	assert.Equal(t, 4, len(items))

	// Sort to ensure we have a consistent order
	sort.Slice(items, func(i, j int) bool {
		return items[i].Id < items[j].Id
	})

	// Check that we have ID 1-4
	for i := 0; i < 4; i++ {
		assert.Equal(t, i+1, items[i].Id)
	}
}

func TestUpdateItem(t *testing.T) {
	// Create and add an item
	item := db.ToDoItem{
		Id:     1234,
		Title:  "A Really Cool Title",
		IsDone: false,
	}
	assert.NoError(t, DB.AddItem(item), "Error adding item to DB")

	// Ensure item is in DB
	dbItem, getErr := DB.GetItem(item.Id)
	assert.NoError(t, getErr, "Error getting item from DB")
	assert.Equal(t, item, dbItem)

	// Update item
	item.Title = "An Even Cooler Title"
	item.IsDone = true
	DB.UpdateItem(item)

	// Get same ID from DB, ensure it matches the updated item
	dbItem, getErr = DB.GetItem(item.Id)
	assert.NoError(t, getErr, "Error getting item from DB")
	assert.Equal(t, item, dbItem)
}

func TestDeleteItem(t *testing.T) {
	// Create a random item
	item := db.ToDoItem{}
	err := fake.Struct(&item)
	assert.NoError(t, err, "Error creating random ToDoItem")

	// Add the item to the DB
	addErr := DB.AddItem(item)
	assert.NoError(t, addErr, "Error adding item to DB")

	// Ensure item was added to DB
	dbItem, getErr := DB.GetItem(item.Id)
	assert.NoError(t, getErr, "Error getting item from DB")
	assert.Equal(t, item, dbItem)

	// Delete item
	deleteErr := DB.DeleteItem(item.Id)
	assert.NoError(t, deleteErr, "Error deleting item from DB")

	// Ensure item doesn't exist in DB
	_, expectedGetErr := DB.GetItem(item.Id)
	assert.Error(t, expectedGetErr, "No error, item must still be in DB")
}

func TestUpdateDoneStatus(t *testing.T) {
	// Create and add an item
	item := db.ToDoItem{
		Id:     4321,
		Title:  "A Really Cool Title",
		IsDone: false,
	}
	assert.NoError(t, DB.AddItem(item), "Error adding item to DB")

	// Ensure item is in DB
	dbItem, getErr := DB.GetItem(item.Id)
	assert.NoError(t, getErr, "Error getting item from DB")
	assert.Equal(t, item, dbItem)

	// Update done status
	assert.NoError(t, DB.ChangeItemDoneStatus(item.Id, true), "Error updating done status")

	// Get same ID from DB, ensure IsDone is true
	dbItem, getErr = DB.GetItem(item.Id)
	assert.NoError(t, getErr, "Error getting item from DB")
	assert.True(t, dbItem.IsDone, "IsDone should be true")

	// Set done status to false
	assert.NoError(t, DB.ChangeItemDoneStatus(item.Id, false), "Error updating done status")

	// Ensure IsDone is false
	dbItem, getErr = DB.GetItem(item.Id)
	assert.NoError(t, getErr, "Error getting item from DB")
	assert.False(t, dbItem.IsDone, "IsDone should be false")
}

func TestRestoreDB(t *testing.T) {
	// Add 5 random objects
	for i := 0; i < 5; i++ {
		item := db.ToDoItem{
			Id:     1000 + i,
			Title:  fake.JobTitle(),
			IsDone: fake.Bool(),
		}
		DB.AddItem(item)
	}

	// Get all items, check that there are at least 9 items in the db
	// (Other tests may add items, so there may be more than 9 items)
	dbItems, getAllErr := DB.GetAllItems()
	assert.NoError(t, getAllErr, "Error getting all items from DB")
	assert.GreaterOrEqual(t, len(dbItems), 9)

	// Restore the DB
	assert.NoError(t, DB.RestoreDB(), "Error restoring DB")

	// Reload the DB from file again
	testdb, newDbErr := db.New(DEFAULT_DB_FILE_NAME)
	assert.NoError(t, newDbErr, "Error loading DB from file")

	// Get all items again, ensure there are only 4 items
	dbItems2, getAllErr2 := testdb.GetAllItems()
	assert.NoError(t, getAllErr2, "Error getting all items from DB")
	assert.Equal(t, 4, len(dbItems2))

}
