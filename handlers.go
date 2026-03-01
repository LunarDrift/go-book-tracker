package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type App struct {
	Library []Book
	Scanner *bufio.Scanner
	Writer  *tabwriter.Writer
}

// Sorting helper:
// Sort by title. If same title, sort by Author
func sortLibrary(books []Book) {
	sort.Slice(books, func(i, j int) bool {
		iTitle := strings.ToLower(books[i].Title)
		jTitle := strings.ToLower(books[j].Title)

		if !strings.EqualFold(iTitle, jTitle) {
			return iTitle < jTitle
		}
		// Check Authors if Titles match
		return strings.ToLower(books[i].Author) < strings.ToLower(books[j].Author)
	})
}

// -----------------------
// Drawing
// -----------------------
func drawBorder() {
	fmt.Printf("------------------------------------------------------------\n")
}

func drawMenu() {
	fmt.Println("\n1. Add Book")
	fmt.Println("2. List Books")
	fmt.Println("3. Remove Book")
	fmt.Println("4. Search")
	fmt.Println("5. Update Status")
	fmt.Println("6. View Stats")
	fmt.Println("0. Exit")
	fmt.Print("Choose an option: ")
}

// -----------------------
// Handler Functions
// -----------------------
func (a *App) handleAddBook() {
	fmt.Print("Enter Title: ")
	a.Scanner.Scan()
	titleInput := a.Scanner.Text()

	fmt.Print("Enter Author: ")
	a.Scanner.Scan()
	authorInput := a.Scanner.Text()

	fmt.Print("Select Status (0: Unread | 1: Reading | 2: Read): ")
	a.Scanner.Scan()
	statusInput, err := strconv.Atoi(a.Scanner.Text())
	// Validate input was a number AND valid status
	if err != nil || statusInput < 0 || statusInput > 2 {
		fmt.Println("Invalid status. Defaulting to Unread.")
		statusInput = 0
	}
	newBook := Book{Title: titleInput, Author: authorInput, Status: ReadStatus(statusInput)}

	// REMEMBER: Re-assign the result of append to the slice!
	a.Library = append(a.Library, newBook)
	sortLibrary(a.Library)
	saveBooks(a.Library)
	fmt.Println("Book added successfully!")
}

func (a *App) handleListBooks() {
	if len(a.Library) == 0 {
		fmt.Println("Your library is empty")
		return
	}

	fmt.Print("\n")
	drawBorder()
	_, _ = fmt.Fprintln(a.Writer, "ID\tTITLE\tAUTHOR\tSTATUS")
	_, _ = fmt.Fprintln(a.Writer, "---\t----------\t----------\t------")
	for i, book := range a.Library {
		_, _ = fmt.Fprintf(a.Writer, "%d. \t%s\t%s\t%s\n", i+1, book.Title, book.Author, book.Status.String())
	}
	if err := a.Writer.Flush(); err != nil {
		// Use Stderr for error messages so they don't get mixed with data
		fmt.Fprintf(os.Stderr, "Error: failed to display search results: %v", err)
		return
	}
	drawBorder()
}

func (a *App) handleRemoveBook() {
	fmt.Print("Enter the Book ID to Remove:")
	a.Scanner.Scan()
	id := a.Scanner.Text()
	// Convert to int
	inputIdx, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error: Please enter a valid numeric ID.")
		return
	}
	// Adjust for 0-based indexing
	actualIdx := inputIdx - 1

	// Safety check: Does this index actually exist?
	if actualIdx < 0 || actualIdx >= len(a.Library) {
		fmt.Printf("Error: Book ID %d does not exist.\n", inputIdx)
		return
	}
	// Removal
	a.Library = append(a.Library[:actualIdx], a.Library[actualIdx+1:]...)
	saveBooks(a.Library)
	fmt.Println("\nBook removed successfully")
}

func (a *App) handleSearchTitle() {
	fmt.Print("Search for a title: ")
	a.Scanner.Scan()
	searchTerm := strings.ToLower(a.Scanner.Text())

	// Find matches first, don't touch writers
	var results []Book
	for _, b := range a.Library {
		if strings.Contains(strings.ToLower(b.Title), searchTerm) {
			results = append(results, b)
		}
	}

	// Now we know what we have before writing anything
	if len(results) == 0 {
		fmt.Println("No books found matching that title.")
		return
	}

	_, _ = fmt.Fprintln(a.Writer, "\nTITLE\tAUTHOR\tSTATUS")
	_, _ = fmt.Fprintln(a.Writer, "---------------\t-------------\t------")
	for _, b := range results {
		_, _ = fmt.Fprintf(a.Writer, "%s\t%s\t%s\n", b.Title, b.Author, b.Status.String())
	}
	if err := a.Writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "error displaying results: %v\n", err)
	}
}

func (a *App) handleSearchAuthor() {
	fmt.Print("Enter Author Name: ")
	a.Scanner.Scan()
	// TrimSpace handles accidental extra spaces
	query := strings.ToLower(strings.TrimSpace(a.Scanner.Text()))

	if query == "" {
		fmt.Println("Search cancelled: No name entered.")
		return
	}

	// Split search into individual words
	searchWords := strings.Fields(query)

	// Create a slice to hold matches
	var results []Book
	for _, b := range a.Library {
		author := strings.ToLower(b.Author)

		// Assume it's a match until proven otherwise
		matchAll := true
		for _, word := range searchWords {
			if !strings.Contains(author, word) {
				matchAll = false
				break
			}
		}

		if matchAll {
			results = append(results, b)
		}
	}

	// Check if we found anything
	if len(results) == 0 {
		fmt.Println("No books found for that author.")
		return
	}

	// Print results
	drawBorder()
	fmt.Printf("\nFound %d matching books:\n", len(results))
	_, _ = fmt.Fprintln(a.Writer, "TITLE\tAUTHOR\tSTATUS")
	_, _ = fmt.Fprintln(a.Writer, "---------------\t-------------\t------")
	for _, b := range results {
		_, _ = fmt.Fprintf(a.Writer, "%s\t%s\t%s\n", b.Title, b.Author, b.Status.String())
	}
	drawBorder()

	if err := a.Writer.Flush(); err != nil {
		// Use Stderr for error messages so they don't get mixed with data
		fmt.Fprintf(os.Stderr, "Error: failed to display search results: %v", err)
	}
}

func (a *App) handleSearchSubMenu() {
	fmt.Print("(1. Search by Title | 2. Search by Author): ")
	a.Scanner.Scan()
	searchInput := a.Scanner.Text()

	switch searchInput {
	case "1":
		a.handleSearchTitle()
	case "2":
		a.handleSearchAuthor()
	default:
		fmt.Println("Invalid option, try again.")
	}
}

func (a *App) handleUpdateStatus() {
	fmt.Print("Enter the Book ID to update status: ")
	a.Scanner.Scan()

	// Validate ID input
	inputIdx, err := strconv.Atoi(a.Scanner.Text())
	if err != nil {
		fmt.Println("Error: Please enter a valid numeric ID.")
	}

	actualIdx := inputIdx - 1
	if actualIdx < 0 || actualIdx >= len(a.Library) {
		fmt.Printf("Error: Book ID %d does not exist.", inputIdx)
	}

	// Get new status
	fmt.Printf("Updating status for %s...\n", a.Library[actualIdx].Title)
	fmt.Print("Select New Status (0: Unread, 1: Reading, 2: Read): ")
	a.Scanner.Scan()

	statusVal, err := strconv.Atoi(a.Scanner.Text())
	if err != nil || statusVal < 0 || statusVal > 2 {
		fmt.Println("Invalid status choice. Update cancelled.")
	}

	// Update the field in memory
	a.Library[actualIdx].Status = ReadStatus(statusVal)

	saveBooks(a.Library)
	fmt.Println("Status updated successfully!")
}

func (a *App) handleStats() {
	if len(a.Library) == 0 {
		fmt.Println("No stats available for an empty library.")
		return
	}

	var unread, reading, read int

	// Map to track "Top Owned Author"
	authorCounts := make(map[string]int)
	topAuthor := "None"
	maxBooks := 0

	for _, book := range a.Library {
		// Add book / increase count in map
		authorCounts[book.Author]++

		// Read/Unread Percentage tracking
		switch book.Status {
		case Unread:
			unread++
		case Reading:
			reading++
		case Read:
			read++
		}
	}
	// Find the "Top Author"
	for author, count := range authorCounts {
		if count > maxBooks {
			maxBooks = count
			topAuthor = author
		}
	}

	drawBorder()
	fmt.Printf("                 Library Summary\n")
	drawBorder()
	fmt.Printf("Total Books:         %d\n", len(a.Library))
	fmt.Printf("Read:                %d\n", read)
	fmt.Printf("Reading:             %d\n", reading)
	fmt.Printf("Unread:              %d\n", unread)

	// Progress calculation
	progress := (float64(read) / float64(len(a.Library))) * 100
	fmt.Printf("Completion:          %.1f%%\n", progress)
	drawBorder()

	fmt.Printf("Your favorite author is %s with %d books!\n", topAuthor, maxBooks)
}

func (a *App) handleExit() {
	fmt.Println("Goodbye!")
	saveBooks(a.Library)
}
