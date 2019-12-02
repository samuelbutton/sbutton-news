package main

import (
	"net/http"
	"os"
	"html/template"
	"net/url"
	"fmt"
	"time"
	"flag"
	"encoding/json"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
)

import . "github.com/scbutton95/news-app/pkg/model"

var apiKey *string
var t *template.Template

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

type Results struct {
	Category     string    `json:"category"`
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

type Search struct {
	SearchKey  string
	NextPage   int
	TotalPages int
	Results    Results
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	digest := make(map[string]Results)
	categories := []string{"business", "general", "health", "science", "sports", "technology"}

	pageSize := 5
	
	var wg sync.WaitGroup
	for i, _ := range categories {
		category := categories[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			currentResults := &Results{}
			endpoint :=	fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=us&category=%s&pageSize=%d&apiKey=%s", category, pageSize, *apiKey)
			
			var issue bool
			w, issue = currentResults.hitAPI(w, endpoint)
			if issue {
				return
			}

			currentResults.Category = strings.Title(category)
			digest[category] = *currentResults
		}()
	}

	wg.Wait()

	err := t.ExecuteTemplate(w, "index.html", digest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("problem with template execution")
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}	

	params := u.Query()
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	search := &Search{}
	search.SearchKey = searchKey

	next, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Unexpected server error", http.StatusInternalServerError)
		return
	}
	search.NextPage = next
	pageSize := 20
	sortBy := "publishedAt"
	language := "en"
	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?&q=%s&pageSize=%d&page=%d&sortBy=%s&language=%s&apiKey=%s", url.QueryEscape(search.SearchKey), pageSize, search.NextPage, sortBy, language, *apiKey)

	var issue bool
	w, issue = search.Results.hitAPI(w, endpoint)
	if issue {
		return
	}

	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))

	if ok := !search.IsLastPage(); ok {
		search.NextPage++
	}

	err = t.ExecuteTemplate(w, "search.html", search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("problem with template execution")
	}
}

func (results *Results) hitAPI(w http.ResponseWriter, endpoint string) (http.ResponseWriter, bool) {
	resp, err := http.Get(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("problem with endpoint pull")
		return w, true
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("problem with response close")
		return w, true
	}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("problem with decode")
		return w, true
	}
	return w, false
}

func (a *Article) FormatPublishedDate() string {
	year, month, day := a.PublishedAt.Date()
	return fmt.Sprintf("%v %d, %d", month, day, year)
}

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

func main() {
	var err error
  	t, err = template.ParseGlob("./templates/*")
  	if err != nil {
		log.Fatal("problem with template parsing")
  	}
  	
  	apiString := os.Getenv("NEWS_API_KEY")
	apiKey = flag.String("apikey", apiString, "Newsapi.org access key")
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("apiKey must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+port, mux)
}
