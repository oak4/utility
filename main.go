package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type response struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []article `json:"articles"`
}

type article struct {
	Source      source `json:"source"`
	Author      string `json:"author"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	URLToImage  string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
	Content     string `json:"content"`
}

type source struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getLink(apiKey, category string) string {
	return fmt.Sprintf("http://newsapi.org/v2/top-headlines?country=us&category=%s&apiKey=%s", category, apiKey)
}

func trimStringToLength(st string, size int) string {
	if len(st) < size {
		return st
	}
	return st[0:size]
}

func getArticles(apiKey, category string) response {
	link := getLink(apiKey, category)

	message := response{}
	resp, err := http.Get(link)
	if err != nil {
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(body, &message)
	if err != nil {
		fmt.Println(err)
	}
	return message

}

func main() {
	apiKey := os.Getenv("OAK4NEWSKEY")
	authors := make(map[string]int)
	sources := make(map[string]int)
	medias := make(map[string]int)
	categories := make(map[string]int)
	articles := make(map[article]int)

	cats := []string{
		"business",
		"entertainment",
		"general",
		"health",
		"science",
		"sports",
		"technology",
	}
	for i, v := range cats {
		categories[v] = i + 1
	}

	f, err := os.Create("insert.sql")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprintln(f, "set define off;")
	fmt.Fprintln(f, "INSERT INTO STATUS VALUES(1,'ACTIV');")
	for key, value := range categories {

		m := getArticles(apiKey, key)

		fmt.Fprintf(f, "INSERT INTO CATEGORII VALUES(%d,'%s',1);\n", value, key)
		for _, v := range m.Articles {
			if len(v.Author) == 0 {
				continue
			}
			var found bool
			_, found = authors[v.Author]
			if !found {
				authors[v.Author] = len(authors) + 1
				fmt.Fprintf(f, "INSERT INTO AUTORI VALUES(%d,q'[%s]',1);\n", authors[v.Author], v.Author)
			}
			_, found = sources[v.Source.Name]
			if !found {
				sources[v.Source.Name] = len(sources) + 1
				fmt.Fprintf(f, "INSERT INTO SURSE VALUES(%d,1,q'[%s]',q'[%s]');\n", sources[v.Source.Name], v.Source.Name, v.Source.Name)
			}
			_, found = medias[v.URLToImage]
			if !found {
				medias[v.URLToImage] = len(medias) + 1
				fmt.Fprintf(f, "INSERT INTO MEDIA VALUES(%d,q'[%s]');\n", medias[v.URLToImage], v.URLToImage)
			}
			_, found = articles[v]
			if !found {
				articles[v] = len(articles) + 1
				fmt.Fprintf(f, "INSERT INTO ARTICOLE VALUES(%d,%d,%d,1,%d,sysdate,q'[%s]',q'[%s]');\n", articles[v], value, authors[v.Author], sources[v.Source.Name], v.Title, trimStringToLength(v.Content, 100))
				fmt.Fprintf(f, "INSERT INTO ARTICOLE_MEDIA VALUES(%d,%d);\n", articles[v], medias[v.URLToImage])
			}

		}
	}
}
