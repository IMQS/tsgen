package report

import (
	"config"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
)

type TSReport struct {
	ReportFields []string
	Name         string
}

func check(e error) {
	if e != nil {
		panic(e)
		os.Exit(1)
	}
}

func (r *TSReport) AddString(m string) {
	r.ReportFields = append(r.ReportFields, m)
}

func (r *TSReport) Create() {
	if r.ReportFields == nil {
		fmt.Println("Empty report")
		os.Exit(1)
	}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(dir + "/Reports/" + r.Name + ".txt")
	check(err)
	defer f.Close()
	for _, field := range r.ReportFields {
		_, err := f.WriteString(field + "\n")
		check(err)
		//fmt.Printf("writing %d", b)
	}

}

func (r *TSReport) AddStruct(t config.TSProperties) {
	s := reflect.ValueOf(&t).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		//r.AddString(fmt.Sprintf("%d: %s %s = %v", i, typeOfT.Field(i).Name, f.Type(), f.Interface()))
		r.AddString(fmt.Sprintf("%d: %s = %v", i, typeOfT.Field(i).Name, f.Interface()))
	}

}
