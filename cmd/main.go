package main

import (
	"net/http"
	"os"
	"html/template"
	"flag"
	"log"
)

var apiKey *string
var t *template.Template

func main() {

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

  	t, err = template.ParseGlob(path + "/cmd/templates/*")
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
	fs := http.FileServer(http.Dir("cmd/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	// mux.HandleFunc("/profile/{slug}", profileHandler, slug)
	// mux.HandleFunc("/signup/", signupHandler)
	// mux.HandleFunc("/login/", loginHandler)
	http.ListenAndServe(":"+port, mux)
}
