/*
Kood/Jõhvi
groupie-tracker
29.09.2021
*/
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

type Group struct {
	Artists   string
	Locations string
	Dates     string
	Relation  string
}

type Artists struct {
	Id           int
	Image        string
	Name         string
	Members      []string
	MemberString string
	CreationDate int
	FirstAlbum   string
	Locations    string
	ConcertDates string
	Relations    string
}

type Relation struct {
	Id             int
	DatesLocations map[string][]string
}

func main() {
	fmt.Println("Server running @ localhost:8080")
	http.HandleFunc("/", handler)
	if http.ListenAndServe(":8080", nil) != nil {
		log.Fatalf("%v - Internal Server Error", http.StatusInternalServerError)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	template := template.Must(template.ParseFiles("templates/index.html"))
	if r.URL.Path != "/" {
		errorHandler(w, 404, "Page not found")
		return
	}

	switch r.Method {
	case "GET":
		// first, let's get all the necessary links from the main page
		body := getBody("https://groupietrackers.herokuapp.com/api")
		var links Group
		err := json.Unmarshal(body, &links)
		jsonErrCheck(err, &w)

		// then, the entire array of artists
		body = getBody(links.Artists)
		var artists []Artists
		err = json.Unmarshal(body, &artists)
		jsonErrCheck(err, &w)

		// now put all that (+ relations) into this slice and output into the template
		var outputs []Artists
		fillOutputs(artists, &outputs)
		template.Execute(w, outputs)
	default:
		fmt.Fprintln(w, "GET is the only method supported.")
	}
}

func getBody(link string) []byte {
	res, err := http.Get(link)
	errCheck(err)
	body, err := io.ReadAll(res.Body)
	errCheck(err)
	res.Body.Close()
	return body
}

func fillOutputs(artists []Artists, outputs *[]Artists) {
	for i := 1; i <= len(artists); i++ {
		body := getBody(artists[i-1].Relations)
		var relations Relation
		err := json.Unmarshal(body, &relations)
		errCheck(err)

		var output = artists[i-1]
		output.MemberString = strings.Join(artists[i-1].Members, ", ")
		output.Relations = map2string(relations.DatesLocations)

		*outputs = append(*outputs, output)
	}
}

func map2string(m map[string][]string) string {
	var s string
	for key, value := range m {
		s += key + " - " + value[0] + "\n"
	}
	return s
}

func errCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func jsonErrCheck(err error, w *http.ResponseWriter) {
	if err != nil {
		http.Error(*w, "400 - Bad Request. Couldn't parse body.", http.StatusBadRequest)
		log.Fatal(err)
	}
}

type Error struct {
	Code        int
	Explanation string
}

var (
	temp_error = template.Must(template.ParseFiles("./templates/error.html"))
)

func errorHandler(w http.ResponseWriter, Code int, Explanation string) {
	w.WriteHeader(Code)
	temp_error.Execute(w, Error{Code: Code, Explanation: Explanation})
}
