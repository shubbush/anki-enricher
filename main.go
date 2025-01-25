package main

import (
	"anki-enricher/anki"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Server struct {
	client *http.Client
}

const VerformanUrl = "https://www.verbformen.de/?w="
const AnkiUrl = "http://localhost:8765"

func main() {
	mux := http.NewServeMux()
	client := &http.Client{}
	server := Server{client}
	mux.HandleFunc("POST /", server.handleReq)
	log.Fatal(http.ListenAndServe("localhost:8766", mux))
}

func (s *Server) handleReq(w http.ResponseWriter, req *http.Request) {
	ankiReq := parseAnkiRequest(req)
	word := ankiReq.Params.Note.Fields["Front"]
	cardHtml := getWordCard(s.client, word)
	// log.Println(string(resp))
	resp, e := sendToAnki(s.client, ankiReq, cardHtml)
	if e != nil {
		log.Fatal(e)
	}
	// fmt.Println(cardHtml)
	// w.Write([]byte("OK"))
	w.Write(resp)
}

func sendToAnki(client *http.Client, ankiReq anki.AddNoteRequest, cardHtml string) ([]byte, error) {
	ankiReq.Params.Note.Fields["Back"] = cardHtml
	ankiReq.Params.Note.Options["allowDuplicate"] = false
	jsonReq, e := json.Marshal(ankiReq)
	if e != nil {
		log.Fatal(e)
	}
	req, _ := http.NewRequest("POST", AnkiUrl, bytes.NewBuffer(jsonReq))
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	return io.ReadAll(res.Body)
}

func parseAnkiRequest(req *http.Request) anki.AddNoteRequest {
	decoder := json.NewDecoder(req.Body)
	var addNoteReq anki.AddNoteRequest
	err := decoder.Decode(&addNoteReq)
	if err != nil {
		log.Fatalf("can't parse request %e", err)
	}
	return addNoteReq
}

func getWordCard(client *http.Client, word string) string {
	doc := requestDocument(client, VerformanUrl+word)
	section := doc.Find("article section").First()
	card := section.Find("div div:nth-child(2)")

	if card.Nodes == nil {
		card = doc.Find("#")
	}

	ruTranslation := doc.Find("dd[lang=ru] span:nth-child(2)")

	if len(ruTranslation.Nodes) > 0 {
		card.Find("span[lang=en]").SetText(ruTranslation.Text())
	}

	card.Find("a").Remove()
	card.Find("#vStckLngImg").Remove()
	card.Find("#vVdPlay").Remove()
	card.Find(".rKnpf").Remove()
	card.Find("#stammformen a").Remove()
	card.Find("a[title='Bedeutungen']").Remove()
	card.Find("img").Remove()

	// card.Find("*").RemoveAttr("class")
	card.Find("*").RemoveAttr("title")
	card.Find("*").RemoveAttr("onclick")
	card.Find("*").RemoveAttr("onchange")
	card.Find("*").RemoveAttr("rel")
	cardHtml, e := card.Html()
	if e != nil {
		log.Fatalf("Can't get card's html %e", e)
	}
	// fmt.Printf("Found card: %s \n", cardHtml)
	return cardHtml
}

func requestDocument(client *http.Client, url string) *goquery.Document {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:133.0) Gecko/20100101 Firefox/133.0")
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}
