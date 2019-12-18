
package main

import (
	"net/http"
	"fmt"
	"encoding/json"

)

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