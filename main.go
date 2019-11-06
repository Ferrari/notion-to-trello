package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/adlio/trello"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type cardData struct {
	title    string
	status   string
	contents string
}

func (card cardData) String() string {
	return fmt.Sprintf("Title: %s\nStatus: %s\nContents: %s\n", card.title, card.status, card.contents)
}

var currentCard cardData

func check(e error) {
	if e != nil {
		panic(e)
	}
}

const statusPrefix = "Status: "

func visitor(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	// If we found a heading and we don't have a title yet, that's our title.
	if len(currentCard.title) == 0 && node.Type == blackfriday.Heading {
		firstChild := *node.FirstChild
		currentCard.title = string(firstChild.Literal)
		return blackfriday.SkipChildren
	}
	if len(currentCard.contents) == 0 && node.Type == blackfriday.Text {
		currentCard.contents = string(node.Literal)
		// Make a scanner to find the status line.
		scanner := bufio.NewScanner(strings.NewReader(currentCard.contents))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, statusPrefix) {
				currentCard.status = strings.TrimSpace(strings.TrimPrefix(line, statusPrefix))
			}
		}
		return blackfriday.SkipChildren
	}
	// Nothing to see here, move along...
	return blackfriday.GoToNext
}

func main() {
	appKey := os.Getenv("TRELLO_KEY")
	token := os.Getenv("TRELLO_TOKEN")
	boardID := os.Args[1]
	if len(boardID) == 0 {
		panic("boardID arg is required")
	}
	todoDir := os.Args[2]
	if len(todoDir) == 0 {
		panic("todoDir is required")
	}

	filenames, err := ioutil.ReadDir(todoDir)
	check(err)
	var cards []cardData
	for _, filename := range filenames {
		contents, err := ioutil.ReadFile(path.Join(todoDir, filename.Name()))
		check(err)
		parser := blackfriday.New()
		parsed := parser.Parse(contents)
		currentCard = cardData{}
		parsed.Walk(visitor)
		cards = append(cards, currentCard)
	}

	client := trello.NewClient(appKey, token)
	log.Println("About to retrieve board...")
	board, err := client.GetBoard(boardID, trello.Defaults())
	check(err)

	log.Println("About to retrieve lists...")
	lists, err := board.GetLists(trello.Defaults())
	check(err)

	var listMap map[string]*trello.List
	listMap = make(map[string]*trello.List)
	for _, list := range lists {
		listMap[list.Name] = list
	}
	log.Println(listMap)

	for _, card := range cards {
		list := listMap[card.status]
		if list == nil {
			panic("Unknown status: " + card.status)
		}
		trelloCard := trello.Card{
			Name: card.title,
			Desc: card.contents,
		}
		log.Printf("Adding %q...\n", card.title)
		list.AddCard(&trelloCard, trello.Defaults())
	}

}
