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
	Bericht string  `json:"nieuwsbericht" xml:"link"`
	Page    string  `json:"subjectpage" xml:"guid"`
	Title   string  `json:"titel" xml:"title"`
	Content string  `json:"inhoud" xml:"description"`
	Date    RSSDate `json:"publicatiedatum" xml:"pubDate"`
}

type Channel struct {
	XMLName     xml.Name        `xml:"channel"`
	Title       string          `xml:"title"`
	Link        string          `xml:"link"`
	Description string          `xml:"description"`
	PubDate     RSSDate         `xml:"pubDate"`
	WebMaster   string          `xml:"webMaster"`
	Items       []Nieuwsbericht `xml:"item"`
}

type RSSDate time.Time

func (d *RSSDate) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*d = RSSDate(t)
	return nil
}

func (d RSSDate) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	formatted := time.Time(d).Format(time.RFC1123Z) // Use RFC1123Z for time zone offset
	return e.EncodeElement(formatted, start)
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

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
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
		PubDate:     RSSDate(time.Now()),
		WebMaster:   "kayasem.vancauwenberghe@ugent.be",
		Items:       make([]Nieuwsbericht, len(items)),
	}

	for i, item := range items {
		item.Date = RSSDate(item.Date)
		feed.Items[i] = item
	}

	rss := RSS{
		Version: "2.0",
		Channel: feed,
	}

	xmlData, err := xml.MarshalIndent(rss, "", "    ")
	if err != nil {
		return nil, err
	}

	rssFeed := []byte(xml.Header + string(xmlData))

	return rssFeed, nil
}
