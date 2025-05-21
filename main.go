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

const (
	URL  = "https://data.stad.gent/api/explore/v2.1/catalog/datasets/recente-nieuwsberichten-van-stadgent/records?limit=100"
	PORT = "8080"
)

func main() {
	fmt.Println("Starting RSS feed generator...")
	fmt.Println("Initial feed generation...")

	if err := generateAndSaveFeed(); err != nil {
		fmt.Println("Error in initial feed generation:", err)
		return
	}

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	go func() {
		http.HandleFunc("/feed", handleFeedRequest)
		fmt.Printf("Starting HTTP server on port %s...\n", PORT)
		if err := http.ListenAndServe(":"+PORT, nil); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	fmt.Println("Feed generator running. Updates every hour...")

	for {
		select {
		case <-ticker.C:
			fmt.Printf("Updating feed at %s...\n", time.Now().Format(time.RFC1123))
			if err := generateAndSaveFeed(); err != nil {
				fmt.Println("Error updating feed:", err)
			} else {
				fmt.Println("Feed updated successfully")
			}
		}
	}
}

func handleFeedRequest(w http.ResponseWriter, r *http.Request) {
	feed, err := os.ReadFile("feed.xml")
	if err != nil {
		http.Error(w, "Error reading feed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write(feed)
}

func generateAndSaveFeed() error {
	rssFeed, err := generateRSSFeed()
	if err != nil {
		return fmt.Errorf("error generating RSS feed: %w", err)
	}

	file, err := os.Create("feed.xml")
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(rssFeed); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
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
