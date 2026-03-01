package main

import (
	"bufio"
	"fmt"
	"os"
	"text/tabwriter"
)

func main() {
	// Initialize "database" (slice of Book structs)
	library, err := loadBooks()
	if err != nil {
		fmt.Printf("Critical Error: %v\n", err)
		os.Exit(1)
	}
	sortLibrary(library)

	// Create a scanner for user input (better than fmt.Scanln for spaces in titles)
	scanner := bufio.NewScanner(os.Stdin)
	// Initialize tabwriter
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 3, ' ', 0)

	app := App{
		Library: library,
		Scanner: scanner,
		Writer:  writer,
	}

	ui := newUI(&app)
	if err := ui.App.Run(); err != nil {
		panic(err)
	}

	// fmt.Println("\n----- Personal Book Tracker -----")
	//
	// for {
	// 	drawMenu()
	// 	scanner.Scan()
	// 	choice := scanner.Text()
	//
	// 	switch choice {
	// 	case "1":
	// 		app.handleAddBook()
	//
	// 	case "2":
	// 		app.handleListBooks()
	//
	// 	case "3":
	// 		app.handleRemoveBook()
	//
	// 	case "4":
	// 		app.handleSearchSubMenu()
	//
	// 	case "5":
	// 		app.handleUpdateStatus()
	//
	// 	case "6":
	// 		app.handleStats()
	//
	// 	case "0":
	// 		app.handleExit()
	// 		return
	//
	// 	default:
	// 		fmt.Println("Invalid option, try again.")
	// 	}
	// }
}
