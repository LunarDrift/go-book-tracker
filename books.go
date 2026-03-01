package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Book struct {
	Title  string     `json:"title"`
	Author string     `json:"author"`
	Status ReadStatus `json:"status"`
}

// Go doesn't have a formal enum keyword. Instead, we create a new type
// based on an int and use iota to assign values

type ReadStatus int

const (
	Unread  ReadStatus = iota // 0
	Reading                   // 1
	Read                      // 2
)

func (s ReadStatus) String() string {
	// A slice of strings where the index matches the constant value
	names := []string{"Unread", "Reading", "Read"}

	// Safety check to prevent "out of bounds" if the int is weird
	if s < Unread || s > Read {
		return "Unknown"
	}
	return names[s]
}

func loadBooks() ([]Book, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dataPath := filepath.Join(homeDir, ".book-tracker", "books.json")

	// 1. Read the file
	data, err := os.ReadFile(dataPath)
	if err != nil {
		// If file doesn't exist, return an empty slice
		return []Book{}, nil
	}

	// 2. Prepare a variable to hold the data
	var loadedBooks []Book

	// 3. Parse the JSON bytes into the slice
	err = json.Unmarshal(data, &loadedBooks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse library file: %w", err)
	}

	// fmt.Println("Library Loaded Successfully")
	return loadedBooks, nil
}

func saveBooks(books []Book) {
	// 1. Turn the slice into "Pretty" JSON bytes
	data, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON: ", err)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dataPath := filepath.Join(homeDir, ".book-tracker", "books.json")

	// 2. Write those bytes to a file
	err = os.WriteFile(dataPath, data, 0o644)
	if err != nil {
		fmt.Println("Error writing to file: ", err)
	}
	// fmt.Println("Library Saved")
}
