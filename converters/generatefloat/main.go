package main

/*  This program generate tests values for oracle NUMBER type

 */

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"strings"
	"text/template"

	_ "github.com/sijms/go-ora"
)

var testValues = []struct {
	asString string
	asFloat  float64
}{
	{"0", 0},
	{"1", 1},
	{"10", 10},
	{"100", 100},
	{"1000", 1000},
	{"10000000", 10000000},
	{"1E+30", 1e+30},
	{"1E+125", 1e+125},
	{"0.1", 0.1},
	{"0.01", 0.01},
	{"0.001", 0.001},
	{"0.0001", 0.0001},
	{"0.00001", 0.00001},
	{"0.000001", 0.000001},
	{"1E+125", 1e125},
	{"1E-125", 1e-125},
	{"-1E+125", -1e125},
	{"-1E-125", -1e-125},
	{"1.23456789e15", 1.23456789e+15},
	{"1.23456789e-15", 1.23456789e-15},
	{"1.234", 1.234},
	{"12.34", 12.34},
	{"123.4", 123.4},
	{"1234", 1234},
	{"12340", 12340},
	{"123400", 123400},
	{"1234000", 1234000},
	{"12340000", 12340000},
	{"0.1234", 0.1234},
	{"0.01234", 0.01234},
	{"0.001234", 0.001234},
	{"0.0001234", 0.0001234},
	{"0.00001234", 0.00001234},
	{"0.000001234", 0.000001234},
	{"-1.234", -1.234},
	{"-12.34", -12.34},
	{"-123.4", -123.4},
	{"-1234", -1234},
	{"-12340", -12340},
	{"-123400", -123400},
	{"-1234000", -1234000},
	{"-12340000", -12340000},
	{"-0.1234", -0.1234},
	{"-1.234", -1.234},
	{"-12.34", -12.34},
	{"-123.4", -123.4},
	{"-1234", -1234},
	{"-12340", -12340},
	{"-123400", -123400},
	{"-1234000", -1234000},
	{"9.8765", 9.8765},
	{"98.765", 98.765},
	{"987.65", 987.65},
	{"9876.5", 9876.5},
	{"98765", 98765},
	{"987650", 987650},
	{"9876500", 9876500},
	{"0.98765", 0.98765},
	{"0.098765", 0.098765},
	{"0.0098765", 0.0098765},
	{"0.00098765", 0.00098765},
	{"0.000098765", 0.000098765},
	{"0.0000098765", 0.0000098765},
	{"0.00000098765", 0.00000098765},
	{"-9.8765", -9.8765},
	{"-98.765", -98.765},
	{"-987.65", -987.65},
	{"-9876.5", -9876.5},
	{"-98765", -98765},
	{"-987650", -987650},
	{"-9876500", -9876500},
	{"-98765000", -98765000},
	{"-0.98765", -0.98765},
	{"-0.098765", -0.098765},
	{"-0.0098765", -0.0098765},
	{"-0.00098765", -0.00098765},
	{"-0.000098765", -0.000098765},
	{"-0.0000098765", -0.0000098765},
	{"-0.00000098765", -0.00000098765},
	{"2*asin(1)", 2. * math.Asin(1.0)},
	{"1/3", 1.0 / 3.0},
	{"-1/3", -1.0 / 3.0},
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type tmplRow struct {
	SelectText string
	OracleText string
	Float      float64
	Binary     string
}

func main() {
	if len(os.Args) < 2 {
		checkErr(fmt.Errorf("Give package name for generated file on command line"))
	}
	packageName := os.Args[1]

	connStr := os.Getenv("GOORA_TESTDB")
	if connStr == "" {
		checkErr(fmt.Errorf("Provide  oracle server url in environment variable GOORA_TESTDB"))
	}
	conn, err := sql.Open("oracle", connStr)
	checkErr(err)
	defer conn.Close()

	result := []tmplRow{}

	for _, tt := range testValues {
		query := fmt.Sprintf("select N||'' S, N, dump(n) D from (select %s N from DUAL)", tt.asString)
		stmt, err := conn.Prepare(query)
		checkErr(err)

		fmt.Println(query)
		rows, err := stmt.Query()
		checkErr(err)

		if !rows.Next() {
			checkErr(fmt.Errorf("Query: %s must return a row", query))
		}
		var (
			asString string
			asFloat  float64
			dump     string
		)

		err = rows.Scan(&asString, &asFloat, &dump)
		checkErr(err)
		i := strings.Index(dump, ": ")
		fmt.Println(tt.asString, tt.asFloat, dump[i+2:])
		result = append(result, tmplRow{
			tt.asString,
			asString,
			tt.asFloat,
			dump[i+2:],
		})

		rows.Close()
		stmt.Close()
	}

	outFile := "testfloatsvalues.go"
	if len(os.Args) > 3 {
		outFile = os.Args[2]
	}
	out, err := os.Create(outFile)
	checkErr(err)
	defer out.Close()

	tmpltext := `package {{.Package}}

/* This file is generated */

var testFloatVualue = []struct {
	SelectText string
	OracleText string
	Float float64
	Binary []byte
}{
	{{- range .Values}}
	{ "{{.SelectText}}", "{{.OracleText}}", {{printf "%g" .Float}}, []byte{ {{.Binary}} } },  // {{printf "%e" .Float}}
	{{- end }}
}`

	tmpl, err := template.New("master").Parse(tmpltext)
	checkErr(err)

	err = tmpl.Execute(out, struct {
		Package string
		Values  []tmplRow
	}{packageName, result})

}