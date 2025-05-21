package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"io"
	"net/http"
	"time"
)

const URL = "https://data.stad.gent/api/explore/v2.1/catalog/datasets/recente-nieuwsberichten-van-stadgent/records?limit=100"

func main() {

	rssFeed, err := generateRSSFeed()
	if err != nil {
		fmt.Println("Error generating RSS feed:", err)
		return
	}

	file, err := os.Create("feed.xml")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	defer file.Close()

	_, err2 := file.Write(rssFeed)
	if err2 != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Rss generated")

}

type NieuwsResponse struct {
	Total   int             `json:"total_count"`
	Results []Nieuwsbericht `json:"results"`
}

type Nieuwsbericht struct {
	Bericht string    `json:"nieuwsbericht" xml:"guid"`
	Page    string    `json:"subjectpage" xml:"link"`
	Title   string    `json:"titel" xml:"title"`
	Content string    `json:"inhoud" xml:"description"`
	Date    time.Time `json:"publicatiedatum" xml:"pubDate"`
}

type Channel struct {
	Title       string          `xml:"title"`
	Link        string          `xml:"link"`
	Description string          `xml:"description"`
	PubDate     time.Time       `xml:"pubDate"`
	Items       []Nieuwsbericht `xml:"item"`
}

func getNieuwsberichten() ([]Nieuwsbericht, error) {
	client := http.Client{Timeout: time.Second * 5}
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var nieuwsResp NieuwsResponse
	if err := json.Unmarshal(body, &nieuwsResp); err != nil {
		return nil, err
	}

	return nieuwsResp.Results, nil
}

func generateRSSFeed() ([]byte, error) {
	items, err := getNieuwsberichten()
	if err != nil {
		return nil, err
	}

	feed := Channel{
		Title:       "Nieuwsberichten Gent",
		Link:        "https://data.stad.gent/explore/dataset/recente-nieuwsberichten-van-stadgent/api/",
		Description: "Recente nieuwsberichten van stad.gent",
		PubDate:     time.Now(),
		Items:       items,
	}

	xmlData, err := xml.MarshalIndent(feed, "", "    ")
	if err != nil {
		return nil, err
	}

	rssFeed := []byte(xml.Header + string(xmlData))

	return rssFeed, nil

}
