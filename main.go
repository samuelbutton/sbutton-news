// declares that code in the main.go file belongs to the main package
package main

import (
	// imports http helper package, provides HTTP client and server implementations
	"net/http"
	// imports os from the system
	"os"
	// templating from go standard library
	"html/template"
	// url processing package
	"net/url"
	// format package
	"fmt"
	// time package
	"time"
	"flag"
	"encoding/json"
	"log"
	"math"
	"strconv"
)

// tpl is a package level variable that points to a template definition
// basically validates the index.html file
// template.Must means that panic ensues if an error is obtained
var tpl = template.Must(template.ParseFiles("index.html"))

var apiKey *string

// Data Model for the application
// to process json from News API
type Source struct {
	ID   interface{} `json:"id"`
	Name string      `json:"name"`
}

type Article struct {
	Source      Source    `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

// formats PublishedAt field in Article
func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d, %d", month, day, year)
}

type Results struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// represents each search query made by the user
type Search struct {
	// the query itself
	SearchKey  string
	// allows us to page through results
	NextPage   int
	// total number of result pages for the query
	TotalPages int
	// current page of results for the query
	Results    Results
}

// to determine if the last page of results has been reached
func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}
	return s.NextPage - 1
}

func (s *Search) PreviousPage()	 int {
	return s.CurrentPage() - 1 
}

// handler function, always of the signature "func(w http.ResponseWriter, r *http.Request)"
// w parameter is the structure we use to send responses to an HTTP request
// it implements a Write() method which accpets a slice of bytes and the writes the data to the connection as part of an HTTP response
// the r parameter represents the HTTP request received from the client
func indexHandler(w http.ResponseWriter, r *http.Request) {

	// execute the template by providing where we want the output to write, and the data we pass to the template
	tpl.Execute(w, nil)

	// previous version
	// w.Write([]byte("<h1>Hello World!</h1>"))
}

func main() {
	// allows us to deine a string flag
	// first argument is the flag name
	// the second is the default value
	// hte third is the usage description
	apiKey = flag.String("apikey", "", "Newsapi.org access key")
	// parse to actually parse the flags
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("apiKey must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// creates a new HTTP request multiplexer
	// multiplexing is the re-use of established server connections for multiple clients connections
	// each client does not require a new connection with the server
	// therefore, the server does not have to allocate sig. resources to establishing / tearing down TCP connections
	// ie, multiple HTTP requests can be sent and responses can be received asynchronously via a single TCP conncection
	// in an example applied to loading a web page, with synchronous you must wait, or can establish multiple connections in a costly manner
	// multiplexing is the best of both worlds
	mux := http.NewServeMux()


	// instatiate a file server object by passing the directory where static files live
	// accesses objects from the Go package imported
	fs := http.FileServer(http.Dir("assets"))

	// tells our touter to use this file server object for all paths beginning with /assets/
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))


	// registers search
	mux.HandleFunc("/search", searchHandler)

	// registers handler function for the root path "/"
	// second argument is handler function
	mux.HandleFunc("/", indexHandler)
	// starts the server on the given port
	http.ListenAndServe(":"+port, mux)
}

// extracts q and page parameter from the request URL and prints them both to the terminal
func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	// getting information from the template
	params := u.Query()
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	// create new instance of Search
	search := &Search{}
	// set SearchKey field on the instance to the value of the q URL parameter in the HTTP request
	search.SearchKey = searchKey

	// convert the page variable into an integer
	// assign the result to the NextPage field of the search instance
	next, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Unexpected server error", http.StatusInternalServerError)
		return
	}
	search.NextPage = next
	
	// number of results the api will return in its response [0 = 100]
	pageSize := 20
	
	// construct the endpoint and make the get request
	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(search.SearchKey), pageSize, search.NextPage, *apiKey)
	resp, err := http.Get(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// closes the response
	defer resp.Body.Close()

	// get request has status code 200 if everything was done ok
	if resp.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&search.Results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}	

	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))
	
	if ok := !search.IsLastPage(); ok {
		search.NextPage++
	}

	// we pass the search variable, instance, as the data interface
	// allows us to access data from the JSON object in our template
	err = tpl.Execute(w, search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

}
