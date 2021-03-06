package main

import (
	"context"
	"log"
	"math/rand"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo"
)

////////////////////////////////////////////////////////////
// Structs used for the talking with frontend
////////////////////////////////////////////////////////////
type Response struct {
	ValidGame bool                   `json:"valid"` // Valid game id
	Payload   map[string]interface{} `json:"payload"`
}

type Card struct {
	Number int    `json:"number"`
	Color  string `json:"color"`
}

func createClient(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := "usu-devops"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done with
	// defer client.Close()
	return client
}

////////////////////////////////////////////////////////////
// Utility functions used in place of firebase
////////////////////////////////////////////////////////////
func randColor(i int) string {
	switch i {
	case 0:
		return "red"
	case 1:
		return "blue"
	case 2:
		return "green"
	case 3:
		return "yellow"
	}
	return ""
}

////////////////////////////////////////////////////////////
// All the data needed for a simulation of the game
// eventually, this will be replaced with firebase
////////////////////////////////////////////////////////////
var gameID string = ""
var currCard []Card = nil // The cards are much easier to render as a list
var players []string = []string{}
var playerIndex = 0 // Used to iterate through the players
var currPlayer string = ""
var allCards map[string][]Card = make(map[string][]Card) // k: username, v: list of cards
var gameStarted bool = false

////////////////////////////////////////////////////////////
// Utility functions
////////////////////////////////////////////////////////////
func newRandomCard() []Card {
	return []Card{Card{rand.Intn(10), randColor(rand.Intn(4))}}
}

func newPayload(user string) map[string]interface{} { // User will default to "" if not passed
	payload := make(map[string]interface{})

	// Update known variables
	payload["current_card"] = currCard
	payload["current_player"] = currPlayer
	payload["all_players"] = players
	payload["deck"] = allCards[user] // returns nil if currPlayer = "" or user not in allCards
	payload["game_id"] = gameID
	payload["game_over"] = checkForWinner()

	return payload
}

func checkID(id string) bool {
	return id == gameID
}

func contains(arr []string, val string) (int, bool) {
	for i, item := range arr {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

////////////////////////////////////////////////////////////
// These are all of the functions for the game -> essentially public functions
////////////////////////////////////////////////////////////
func updateGame(c echo.Context) *Response {
	success := false
	if success = checkID(c.Param("game")); success && gameStarted {
		return &Response{true, newPayload(c.Param("username"))}
	}
	return &Response{false, nil}
}

func createNewGame(c echo.Context) *Response {

	/*
		ctx := context.Background()
			client := createClient(ctx)

			_, _, err := client.Collection("users").Add(ctx, map[string]interface{}{
				"first": "Ada",
				"last":  "Lovelace",
				"born":  1815,
			})
			if err != nil {
				log.Fatalf("Failed adding alovelace: %v", err)
			}
	*/
	gameID = "12234"
	return &Response{true, newPayload("")}
}

func joinGame(c echo.Context) *Response {
	//ctx := context.Background()
	//client := createClient(ctx)

	// iter := client.Collection("users").Documents(ctx)
	// for {
	// 	doc, err := iter.Next()
	// 	if err == iterator.Done {
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Fatalf("Failed to iterate: %v", err)
	// 	}
	// 	fmt.Println(doc.Data())
	// }

	if checkID(c.Param("game")) {
		user := c.Param("username")
		if _, found := contains(players, user); !found {
			players = append(players, user)
			allCards[user] = nil // No cards yet
		}
		return &Response{true, newPayload(c.Param("username"))}
	}
	return &Response{false, nil} // bad game_id
}

func playCard(c echo.Context, card Card) *Response {
	if checkID(c.Param("game")) && currPlayer == c.Param("username") {
		if card.Color == currCard[0].Color || card.Number == currCard[0].Number {
			// Valid card can be played
			playerIndex = (playerIndex + 1) % len(players)
			currPlayer = players[playerIndex]
			currCard[0] = card

			for index, item := range allCards[c.Param("username")] {
				if item == currCard[0] {
					allCards[c.Param("username")] = append(allCards[c.Param("username")][:index], allCards[c.Param("username")][index+1:]...)
					break
				}
			}
		}
		return &Response{true, newPayload(c.Param("username"))}
	}

	return &Response{false, nil}
}

// TODO: Keep track of current card that is top of the deck
func drawCard(c echo.Context) *Response {
	if checkID(c.Param("game")) && c.Param("username") == currPlayer {
		playerIndex = (playerIndex + 1) % len(players)
		currPlayer = players[playerIndex]
		allCards[c.Param("username")] = append(allCards[c.Param("username")], newRandomCard()[0])
		return &Response{true, newPayload(c.Param("username"))}
	}
	return &Response{false, nil}
}

// TODO: need to deal the actual cards, not just random numbers
func dealCards() {
	// The game has started, no more players are joining
	// loop through players, set their cards
	gameStarted = true
	currPlayer = players[rand.Intn(len(players))]

	for k := range players {
		cards := []Card{}
		for i := 0; i < 7; i++ {
			cards = append(cards, Card{rand.Intn(10), randColor(rand.Intn(4))})
		}
		allCards[players[k]] = cards
	}

	currCard = newRandomCard()
}

// TODO: make sure this reflects on the front end
func checkForWinner() string {
	for k := range players {
		if len(allCards[players[k]]) == 0 {
			return players[k]
		}
	}
	return ""
}
