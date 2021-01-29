package main

// https://pkg.go.dev/github.com/tealeg/xlsx

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/tealeg/xlsx/v3"
)

type Log interface {
	Printf(msg string, args ...interface{})
	Println(args ...interface{})
}

type Server struct {
	log   Log
	views *template.Template
	xls   *xlsx.File
}

func NewServer(logger Log, templatesPath, excelFileName string) (*Server, error) {
	views, err := template.ParseGlob(templatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load templates from %q: %w", templatesPath, err)
	}

	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Opened", excelFileName)

	return &Server{
		log:   logger,
		views: views,
		xls:   xlFile,
	}, nil
}

func weekStartDate(date time.Time) time.Time {
	offset := (int(time.Monday) - int(date.Weekday()) - 7) % 7
	result := date.Add(time.Duration(offset*24) * time.Hour)
	return result
}

func main() {
	excelFileName := flag.String("xlsx", "export.xlsx", "export")
	flag.Parse()
	logger := log.New(os.Stderr, "", log.Lshortfile)

	server, err := NewServer(logger, "templates/*.html", *excelFileName)
	if err != nil {
		logger.Fatalf("failed to create server: %v", err)
	}

	err = http.ListenAndServe(":"+os.Getenv("PORT"), server)
	if err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query().Get("date")
	if len(d) != 10 {
		t := time.Now().Local()
		d = t.Format("2006-01-02")
	}

	date, err := time.Parse(time.RFC3339[0:10], d)
	if err != nil {
		server.log.Printf("Could not parse date: %s", d)
		http.Error(w, "Bad date", http.StatusBadRequest)
		return
	}

	server.log.Println("date", date)

	b, err := server.takeNames(date)
	if err != nil {
		server.log.Printf("Roster date not found: %s", date)
		http.Error(w, "Roster date not found", http.StatusBadRequest)
		return
	}

	server.views.ExecuteTemplate(w, "index.html", struct {
		Date  string
		Names []string
	}{
		Date:  d,
		Names: b,
	})

}

func (server *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(data)
}

func (server *Server) takeNames(date time.Time) (names []string, err error) {

	sheetName := weekStartDate(date).Format("Week starting 2 Jan")

	for _, sheet := range server.xls.Sheets {
		if sheet.Name == sheetName {
			log.Printf("Found sheet: %s", sheetName)
			var x, y int
			sheet.ForEachRow(func(row *xlsx.Row) error {
				row.ForEachCell(func(cell *xlsx.Cell) error {
					value, err := cell.FormattedValue()
					if err == nil {
						//if value != "" {
						//	fmt.Println(value)
						//}
						if value == date.Format(`02\-Jan\-2006`) || value == date.Weekday().String() {
							x, y = cell.GetCoordinates()
							log.Printf("%d %d\n", x, y)
						}
					}
					return err
				})
				return err
			})
			if x == 0 {
				log.Fatalf("Could not find roster for %s", date.Format(`02\-Jan\-2006`))
			}
			return getNames(sheet, y, x)
		}
	}
	return
}

func getNames(sheet *xlsx.Sheet, x, y int) (names []string, err error) {
	// https://github.com/tealeg/xlsx/blob/master/tutorial/tutorial.adoc#working-with-rows-and-cells
	for i := x + 1; i < sheet.MaxRow; i++ {
		theCell, err := sheet.Cell(i, y)
		if err != nil {
			return names, err
		}
		fv, err := theCell.FormattedValue()
		if err != nil {
			return names, err
		}
		if fv != "" {
			names = append(names, fv)
		}
	}
	return
}
