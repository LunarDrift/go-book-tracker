package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// -----------------------
// UI holds all the tview primitives that need to talk to each other.
// Think of it like your App struct — it's the single place that owns
// every widget so callbacks can reference them.
// -----------------------

type UI struct {
	App     *tview.Application
	AppData *App // Your existing App struct (library, scanner, etc.)

	// Layout
	Grid *tview.Grid

	// Panels
	BookTable  *tview.Table
	DetailView *tview.TextView
	StatsView  *tview.TextView
	SearchBar  *tview.InputField
	MenuList   *tview.TextView
	StatusBar  *tview.TextView

	// Modal / form overlay (for Add Book)
	AddBookForm *tview.Form
	Pages       *tview.Pages // Pages lets you layer modals on top of the main layout
}

// -----------------------
// newUI wires everything together and returns a ready-to-run UI.
// Call this from main() instead of your current for-loop.
// -----------------------
func newUI(appData *App) *UI {
	ui := &UI{
		App:     tview.NewApplication(),
		AppData: appData,
	}

	ui.Pages = tview.NewPages()

	ui.buildStatusBar()
	ui.buildSearchBar()
	ui.buildMenuList()
	ui.buildBookTable()
	ui.buildDetailView()
	ui.buildStatsView()
	ui.buildGrid()
	ui.buildAddBookForm()

	// Main layout lives on the "main" page
	ui.Pages.AddPage("main", ui.Grid, true, true)

	// Populate widgets with data from the library
	ui.refreshBookTable()
	ui.refreshStats()

	// Global keybindings
	ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Check if we're in any kind of text input - if so, pass the key through untouched
		focused := ui.App.GetFocus()
		_, isInput := focused.(*tview.InputField)
		_, isTextArea := focused.(*tview.TextArea)
		_, isDropDown := focused.(*tview.DropDown)
		if isInput || isTextArea || isDropDown {
			return event
		}

		// Now it's safe to intercept keys
		switch event.Rune() {
		case 'a':
			ui.showAddBookForm()
			return nil
		case 'r':
			ui.handleRemoveSelected()
			return nil
		case 'u':
			ui.handleUpdateStatusSelected()
			return nil
		case 'q':
			ui.App.Stop()
			return nil
		case '/':
			ui.App.SetFocus(ui.SearchBar)
			return nil
		}
		// Escape returns focus to the book table
		if event.Key() == tcell.KeyEscape {
			ui.Pages.SwitchToPage("main")
			ui.App.SetFocus(ui.BookTable)
			return nil
		}
		return event
	})

	ui.App.SetRoot(ui.Pages, true)
	return ui
}

// -----------------------
// Build functions — one per widget
// -----------------------

func (ui *UI) buildStatusBar() {
	ui.StatusBar = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
}

func (ui *UI) buildSearchBar() {
	ui.SearchBar = tview.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(0) // 0 = fill available width

	ui.SearchBar.SetBorder(true)

	// Fires on every keystroke — live filtering
	ui.SearchBar.SetChangedFunc(func(text string) {
		ui.filterBookTable(text)
	})

	// Enter moves focus to the table so you can navigate results
	ui.SearchBar.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			ui.App.SetFocus(ui.BookTable)
		}
	})
}

func (ui *UI) buildMenuList() {
	ui.MenuList = tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow](a)[white] Add Book\n[yellow](r)[white] Remove Book\n[yellow](u)[white] Update Status\n[yellow](q)[white] Quit")

	ui.MenuList.SetBorder(true).SetTitle("Shortcuts")
	// ui.MenuList = tview.NewList().
	// 	AddItem("Add Book", "", 'a', func() {
	// 		ui.showAddBookForm()
	// 	}).
	// 	AddItem("Remove Book", "Select a book first", 'r', func() {
	// 		ui.handleRemoveSelected()
	// 	}).
	// 	AddItem("Update Status", "Select a book first", 'u', func() {
	// 		ui.handleUpdateStatusSelected()
	// 	}).
	// 	AddItem("Exit", "", 'q', func() {
	// 		ui.App.Stop()
	// 	})
	//
	// ui.MenuList.SetBorder(true).SetTitle("Menu")
}

func (ui *UI) buildBookTable() {
	ui.BookTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false) // selectable rows, not columns

	ui.BookTable.SetBorder(true).SetTitle("Library")

	// When user highlights a row, update the detail panel
	ui.BookTable.SetSelectionChangedFunc(func(row, col int) {
		ui.updateDetailView(row)
	})

	// Enter on a row cycles its read status (quick shortcut)
	ui.BookTable.SetSelectedFunc(func(row, col int) {
		ui.handleCycleStatus(row)
	})
}

func (ui *UI) buildDetailView() {
	ui.DetailView = tview.NewTextView().
		SetDynamicColors(true). // enables [red] style color tags
		SetWrap(true)

	ui.DetailView.SetBorder(true).SetTitle("Detail")
}

func (ui *UI) buildStatsView() {
	ui.StatsView = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false)

	ui.StatsView.SetBorder(true).SetTitle("Stats")
}

func (ui *UI) buildGrid() {
	ui.Grid = tview.NewGrid().
		// 4 rows: status bar (1 line tall), search bar (3 lines tall), main content (flexible), stats (8 lines tall)
		SetRows(1, 3, 0, 10).
		// 2 columns: left sidebar (30 wide), main book list (flexible)
		SetColumns(40, 0).
		SetBorders(false)

	ui.Grid.AddItem(ui.StatusBar, 0, 0, 1, 2, 0, 0, false)

	// Search bar spans both columns at the top
	ui.Grid.AddItem(ui.SearchBar, 1, 0, 1, 2, 0, 0, false)

	// Menu sits top-left in the middle row
	ui.Grid.AddItem(ui.MenuList, 2, 0, 1, 1, 0, 0, false)

	// Stats sits bottom-left
	ui.Grid.AddItem(ui.StatsView, 3, 0, 1, 1, 0, 0, false)

	// Book table spans the right column for both middle and bottom rows
	ui.Grid.AddItem(ui.BookTable, 2, 1, 1, 1, 0, 0, true) // true = has focus on start

	// Detail view sits below the book table (or swap this to span right if preferred)
	ui.Grid.AddItem(ui.DetailView, 3, 1, 1, 1, 0, 0, false)
}

func (ui *UI) buildAddBookForm() {
	ui.AddBookForm = tview.NewForm().
		AddInputField("Title", "", 40, nil, nil).
		AddInputField("Author", "", 40, nil, nil).
		AddDropDown("Status", []string{"Unread", "Reading", "Read"}, 0, nil).
		AddButton("Add", func() {
			ui.submitAddBookForm()
		}).
		AddButton("Cancel", func() {
			ui.Pages.SwitchToPage("main")
			ui.App.SetFocus(ui.BookTable)
		})

	ui.AddBookForm.SetBorder(true).SetTitle("Add New Book")

	// Wrap in a centering grid
	// 0 in SetRows/Columns means "flexible" - it expands to fill space
	// The fixed values (40, 14) are the width and height of the form itself
	centeredForm := tview.NewGrid().
		SetColumns(0, 55, 0).
		SetRows(0, 14, 0).
		AddItem(ui.AddBookForm, 1, 1, 1, 1, 0, 0, true)

	ui.Pages.AddPage("addBook", centeredForm, true, true)
}

// -----------------------
// Data → UI refresh functions
// These are the equivalent of your old handleListBooks / handleStats.
// Instead of printing, they write into tview primitives.
// -----------------------

func (ui *UI) refreshBookTable() {
	ui.BookTable.Clear()

	// Header row
	headers := []string{"#", "Title", "Author", "Status"}
	for col, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false) // header row can't be selected
		ui.BookTable.SetCell(0, col, cell)
	}

	// Data rows
	for i, book := range ui.AppData.Library {
		row := i + 1
		ui.BookTable.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", row)))
		ui.BookTable.SetCell(row, 1, tview.NewTableCell(book.Title))
		ui.BookTable.SetCell(row, 2, tview.NewTableCell(book.Author))
		ui.BookTable.SetCell(row, 3, tview.NewTableCell(book.Status.String()).
			SetTextColor(statusColor(book.Status)))
	}

	ui.refreshStats()
}

func (ui *UI) refreshStats() {
	var unread, reading, read int
	authorCounts := make(map[string]int)

	for _, book := range ui.AppData.Library {
		authorCounts[book.Author]++
		switch book.Status {
		case Unread:
			unread++
		case Reading:
			reading++
		case Read:
			read++
		}
	}

	topAuthor := "None"
	maxBooks := 0
	for author, count := range authorCounts {
		if count > maxBooks {
			maxBooks = count
			topAuthor = author
		}
	}

	total := len(ui.AppData.Library)
	var completion float64
	if total > 0 {
		completion = (float64(read) / float64(total)) * 100
	}

	// tview uses [color] tags for color when SetDynamicColors is true
	stats := fmt.Sprintf(
		"[yellow]Total:[white]   %d\n[green]Read:[white]    %d\n[blue]Reading:[white] %d\n[red]Unread:[white]  %d\n[yellow]Done:[white]    %.1f%%\n\n[white]Top Author:\n%s",
		total, read, reading, unread, completion, topAuthor,
	)

	ui.StatsView.SetText(stats)
}

func (ui *UI) updateDetailView(row int) {
	// Row 0 is the header, so subtract 1 to get library index
	idx := row - 1
	if idx < 0 || idx >= len(ui.AppData.Library) {
		ui.DetailView.SetText("")
		return
	}

	book := ui.AppData.Library[idx]
	detail := fmt.Sprintf(
		"[yellow]Title:[white]\n%s\n\n[yellow]Author:[white]\n%s\n\n[yellow]Status:[white]\n%s",
		book.Title, book.Author, book.Status.String(),
	)
	ui.DetailView.SetText(detail)
}

func (ui *UI) filterBookTable(searchTerm string) {
	ui.BookTable.Clear()

	// Rewrite the header
	headers := []string{"#", "Title", "Author", "Status"}
	for col, h := range headers {
		ui.BookTable.SetCell(0, col, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false))
	}

	row := 1
	term := strings.ToLower(searchTerm)
	for _, book := range ui.AppData.Library {
		if term == "" ||
			strings.Contains(strings.ToLower(book.Title), term) ||
			strings.Contains(strings.ToLower(book.Author), term) {

			ui.BookTable.SetCell(row, 0, tview.NewTableCell(fmt.Sprintf("%d", row)))
			ui.BookTable.SetCell(row, 1, tview.NewTableCell(book.Title))
			ui.BookTable.SetCell(row, 2, tview.NewTableCell(book.Author))
			ui.BookTable.SetCell(row, 3, tview.NewTableCell(book.Status.String()).
				SetTextColor(statusColor(book.Status)))
			row++
		}
	}
}

// -----------------------
// Action handlers
// -----------------------

func (ui *UI) showAddBookForm() {
	// Clear any previous input
	ui.AddBookForm.GetFormItemByLabel("Title").(*tview.InputField).SetText("")
	ui.AddBookForm.GetFormItemByLabel("Author").(*tview.InputField).SetText("")

	ui.Pages.SwitchToPage("addBook")
	ui.App.SetFocus(ui.AddBookForm)
}

func (ui *UI) submitAddBookForm() {
	title := ui.AddBookForm.GetFormItemByLabel("Title").(*tview.InputField).GetText()
	author := ui.AddBookForm.GetFormItemByLabel("Author").(*tview.InputField).GetText()
	_, statusStr := ui.AddBookForm.GetFormItemByLabel("Status").(*tview.DropDown).GetCurrentOption()

	if strings.TrimSpace(title) == "" || strings.TrimSpace(author) == "" {
		// TODO: show an error modal
		return
	}

	statusMap := map[string]ReadStatus{"Unread": Unread, "Reading": Reading, "Read": Read}
	newBook := Book{Title: title, Author: author, Status: statusMap[statusStr]}

	ui.AppData.Library = append(ui.AppData.Library, newBook)
	sortLibrary(ui.AppData.Library)
	saveBooks(ui.AppData.Library)

	ui.Pages.SwitchToPage("main")
	ui.App.SetFocus(ui.BookTable)
	ui.refreshBookTable()
	ui.showMessage("Book added successfully!")
}

func (ui *UI) handleRemoveSelected() {
	row, _ := ui.BookTable.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(ui.AppData.Library) {
		return
	}

	// Show a confirmation modal before deleting
	bookTitle := ui.AppData.Library[idx].Title
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Remove \"%s\"?", bookTitle)).
		AddButtons([]string{"Remove", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Remove" {
				ui.AppData.Library = append(
					ui.AppData.Library[:idx],
					ui.AppData.Library[idx+1:]...,
				)
				saveBooks(ui.AppData.Library)
				ui.refreshBookTable()
				ui.showMessage(fmt.Sprintf("Removed \"%s\"", bookTitle))
			}
			ui.Pages.RemovePage("confirm")
			ui.App.SetFocus(ui.BookTable)
		})

	ui.Pages.AddPage("confirm", modal, true, true)
	ui.App.SetFocus(modal)
}

func (ui *UI) handleUpdateStatusSelected() {
	row, _ := ui.BookTable.GetSelection()
	idx := row - 1
	if idx < 0 || idx >= len(ui.AppData.Library) {
		return
	}

	// Reuse the modal pattern to pick a new status
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Update status for \"%s\"", ui.AppData.Library[idx].Title)).
		AddButtons([]string{"Unread", "Reading", "Read", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			statusMap := map[string]ReadStatus{"Unread": Unread, "Reading": Reading, "Read": Read}
			if newStatus, ok := statusMap[buttonLabel]; ok {
				ui.AppData.Library[idx].Status = newStatus
				saveBooks(ui.AppData.Library)
				ui.refreshBookTable()
			}
			ui.Pages.RemovePage("statusPicker")
			ui.App.SetFocus(ui.BookTable)
		})

	ui.Pages.AddPage("statusPicker", modal, true, true)
	ui.App.SetFocus(modal)
}

// Pressing Enter on a row cycles: Unread → Reading → Read → Unread
func (ui *UI) handleCycleStatus(row int) {
	idx := row - 1
	if idx < 0 || idx >= len(ui.AppData.Library) {
		return
	}
	current := ui.AppData.Library[idx].Status
	ui.AppData.Library[idx].Status = ReadStatus((int(current) + 1) % 3)
	saveBooks(ui.AppData.Library)
	ui.refreshBookTable()
	ui.showMessage(fmt.Sprintf("Status updated to %s", ui.AppData.Library[idx].Status.String()))
}

func (ui *UI) showMessage(msg string) {
	ui.StatusBar.SetText(msg)
	go func() {
		time.Sleep(2 * time.Second)
		ui.App.QueueUpdateDraw(func() {
			ui.StatusBar.SetText("")
		})
	}()
}

// -----------------------
// Helpers
// -----------------------

func statusColor(s ReadStatus) tcell.Color {
	switch s {
	case Read:
		return tcell.ColorGreen
	case Reading:
		return tcell.ColorBlue
	case Unread:
		return tcell.ColorRed
	default:
		return tcell.ColorWhite
	}
}
