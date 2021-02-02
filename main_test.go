package main

import (
	"reflect"
	"testing"
	"text/template"
	"time"

	"github.com/tealeg/xlsx/v3"
)

var specificdate, _ = time.Parse(time.RFC3339[0:10], "2021-02-07")

func TestServer_takeNames(t *testing.T) {
	xlFile, err := xlsx.OpenFile("testdata/test.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	type fields struct {
		log   Log
		views *template.Template
		xls   *xlsx.File
	}
	type args struct {
		date time.Time
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantNames []string
		wantErr   bool
	}{
		{
			name: "Basic",
			fields: fields{
				log:   nil,
				views: nil,
				xls:   xlFile,
			},
			args: args{
				date: specificdate,
			},
			wantNames: []string{"Michael", "Brian", "Elsie", "Ernest"},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				log:   tt.fields.log,
				views: tt.fields.views,
				xls:   tt.fields.xls,
			}
			gotNames, err := server.takeNames(tt.args.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.takeNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotNames, tt.wantNames) {
				t.Errorf("Server.takeNames() = %v, want %v", gotNames, tt.wantNames)
			}
		})
	}
}
