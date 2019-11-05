package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	trello "github.com/adlio/trello"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func visitor(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	log.Println(node.Type, entering, string(node.Literal))
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
		listMap[strings.ToLower(list.Name)] = list
	}
	log.Println(listMap)

	filenames, err := ioutil.ReadDir(todoDir)
	check(err)
	for _, filename := range filenames {
		contents, err := ioutil.ReadFile(path.Join(todoDir, filename.Name()))
		check(err)
		parser := blackfriday.New()
		parsed := parser.Parse(contents)
		parsed.Walk(visitor)
		panic(3)
	}
}
