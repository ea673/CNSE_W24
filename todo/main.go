package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"drexel.edu/todo/db"
)

// Global variables to hold the command line flags to drive the todo CLI
// application
var (
	dbFileNameFlag string
	restoreDbFlag  bool
	listFlag       bool
	itemStatusFlag bool
	queryFlag      int
	addFlag        string
	updateFlag     string
	deleteFlag     int
	rootCmd        = &cobra.Command{
		Use:   "todo",
		Short: "A CLI that keeps track of your ToDo items",
	}
)

type AppOptType int

// To make the code a little more clean we will use the following
// constants as basically enumerations for different options.  This
// allows us to use a switch statement in main to process the command line
// flags effectively
const (
	LIST_DB_ITEM AppOptType = iota
	RESTORE_DB_ITEM
	QUERY_DB_ITEM
	ADD_DB_ITEM
	UPDATE_DB_ITEM
	DELETE_DB_ITEM
	CHANGE_ITEM_STATUS
	NOT_IMPLEMENTED
	INVALID_APP_OPT
)

// processCmdLineFlags parses the command line flags for our CLI
//
// 		 This function uses the flag package to parse the command line
//		 flags.  The flag package is not very flexible and can lead to
//		 some confusing code.

//			 REQUIRED:     Study the code below, and make sure you understand
//						   how it works.  Go online and readup on how the
//						   flag package works.  Then, write a nice comment
//				  		   block to document this function that highights that
//						   you understand how it works.
//
//			 EXTRA CREDIT: The best CLI and command line processor for
//						   go is called Cobra.  Refactor this function to
//						   use it.  See github.com/spf13/cobra for information
//						   on how to use it.
//
//	 YOUR ANSWER: This function uses the flag package to define all the supported
//					command line flags and then parse them from the command line.
//					It then loops over all the set flags in lexicographical order
//					and will return the constant associated with the last flag.
//					If there is an error, it will be returned along with the
//					INVALID_APP_OPT constant.
func processCmdLineFlags() (AppOptType, error) {
	rootCmd.PersistentFlags().StringVar(&dbFileNameFlag, "db", "./data/todo.json", "Name of the database file")
	rootCmd.PersistentFlags().BoolVarP(&restoreDbFlag, "restore", "r", false, "Restore the database from the backup file")
	rootCmd.PersistentFlags().BoolVarP(&listFlag, "list", "l", false, "List all the items in the database")
	rootCmd.PersistentFlags().IntVarP(&queryFlag, "query", "q", 0, "Query an item in the database")
	rootCmd.PersistentFlags().StringVarP(&addFlag, "add", "a", "", "Add an item to the database")
	rootCmd.PersistentFlags().StringVarP(&updateFlag, "update", "u", "", "Update an item in the database")
	rootCmd.PersistentFlags().IntVarP(&deleteFlag, "delete", "d", 0, "Delete an item from the database")
	rootCmd.PersistentFlags().BoolVarP(&itemStatusFlag, "statuschange", "s", false, "Change item 'done' status to true or false. Must be used in conjunction with -q to specify the item.")

	err := rootCmd.Execute()
	if err != nil {
		return INVALID_APP_OPT, err
	}

	var appOpt AppOptType = INVALID_APP_OPT

	//show help if no flags are set
	if len(os.Args) == 1 {
		rootCmd.Usage()
		return appOpt, errors.New("no flags were set")
	}

	// Flag to keep track of whether the -q option was set.
	// Since queryFlag is an int, it will default to 0 so we can't use
	// the value to determine whether the flag was set or not.
	// We could rely on appOpt, however if a user decides to also
	// set the --restore flag, that will overwrite appOpt due to the order
	// in which the flags are visited. Therefore the best solution is to use a flag value
	queryFlagSet := false

	// Loop over the flags and check which ones are set, set appOpt
	// accordingly
	rootCmd.Flags().Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "list":
			appOpt = LIST_DB_ITEM
		case "restore":
			appOpt = RESTORE_DB_ITEM
		case "query":
			queryFlagSet = true
			appOpt = QUERY_DB_ITEM
		case "add":
			appOpt = ADD_DB_ITEM
		case "update":
			appOpt = UPDATE_DB_ITEM
		case "delete":
			appOpt = DELETE_DB_ITEM

		//EXTRA CREDIT - Implment the -s flag that changes the
		//done status of an item in the database.  For example -s=true
		//will set the done status for a particular item to true, and
		//-s=false will set the done states for a particular item to
		//false.
		//
		//HINT FOR EXTRA CREDIT
		//Note the -s option also requires an id for the item to that
		//you want to change.  I recommend you use the -q option to
		//specify the item id.  Therefore, the -s option is only valid
		//if the -q option is also set
		case "statuschange":
			// The -s option is only valid if used with the -q flag
			if queryFlagSet {
				appOpt = CHANGE_ITEM_STATUS
			} else {
				fmt.Println("Item to update not specified. Please specify with -q option.")
				appOpt = INVALID_APP_OPT
			}
		default:
			appOpt = INVALID_APP_OPT
		}
	})

	if appOpt == INVALID_APP_OPT || appOpt == NOT_IMPLEMENTED {
		fmt.Println("Invalid option set or the desired option is not currently implemented")
		return appOpt, errors.New("no flags or unimplemented were set")
	}

	return appOpt, nil
}

// main is the entry point for our todo CLI application.  It processes
// the command line flags and then uses the db package to perform the
// requested operation
func main() {

	//Process the command line flags
	opts, err := processCmdLineFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//Create a new db object
	todo, err := db.New(dbFileNameFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//Switch over the command line flags and call the appropriate
	//function in the db package
	switch opts {
	case RESTORE_DB_ITEM:
		fmt.Println("Running RESTORE_DB_ITEM...")
		if err := todo.RestoreDB(); err != nil {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println("Database restored from backup file")
	case LIST_DB_ITEM:
		fmt.Println("Running QUERY_DB_ITEM...")
		todoList, err := todo.GetAllItems()
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		for _, item := range todoList {
			todo.PrintItem(item)
		}
		fmt.Println("THERE ARE", len(todoList), "ITEMS IN THE DB")
		fmt.Println("Ok")

	case QUERY_DB_ITEM:
		fmt.Println("Running QUERY_DB_ITEM...")
		item, err := todo.GetItem(queryFlag)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		todo.PrintItem(item)
		fmt.Println("Ok")
	case ADD_DB_ITEM:
		fmt.Println("Running ADD_DB_ITEM...")
		item, err := todo.JsonToItem(addFlag)
		if err != nil {
			fmt.Println("Add option requires a valid JSON todo item string")
			fmt.Println("Error: ", err)
			break
		}
		if err := todo.AddItem(item); err != nil {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println("Ok")
	case UPDATE_DB_ITEM:
		fmt.Println("Running UPDATE_DB_ITEM...")
		item, err := todo.JsonToItem(updateFlag)
		if err != nil {
			fmt.Println("Update option requires a valid JSON todo item string")
			fmt.Println("Error: ", err)
			break
		}
		if err := todo.UpdateItem(item); err != nil {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println("Ok")
	case DELETE_DB_ITEM:
		fmt.Println("Running DELETE_DB_ITEM...")
		err := todo.DeleteItem(deleteFlag)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println("Ok")
	case CHANGE_ITEM_STATUS:
		fmt.Println("Running CHANGE_ITEM_STATUS...")
		err := todo.ChangeItemDoneStatus(queryFlag, itemStatusFlag)
		if err != nil {
			fmt.Println("Error: ", err)
			break
		}
		fmt.Println("Ok")
	default:
		fmt.Println("INVALID_APP_OPT")
	}
}
