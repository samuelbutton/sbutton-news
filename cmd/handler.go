package main

import (
	"net/http"
	"net/url"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
)

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

// functionality to be added

func signupHandler(w http.ResponseWriter, r *http.Request) {

}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	
}
func profileHandler(w http.ResponseWriter, r *http.Request) {
	
}