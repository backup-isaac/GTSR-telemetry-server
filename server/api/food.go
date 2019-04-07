package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var foodLock sync.Mutex

type fooddb map[string][]string

// FoodSuggestion handles a request from Slack asking for a suggestion of
// where to eat
func (api *API) FoodSuggestion(res http.ResponseWriter, req *http.Request) {
	foodLock.Lock()
	defer foodLock.Unlock()
	var db fooddb = make(map[string][]string)
	rawJSON, err := ioutil.ReadFile("/fooddb/food.json")
	if err == nil {
		err = json.Unmarshal(rawJSON, &db)
		if err != nil {
			log.Printf("Malformatted JSON in food database")
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	err = req.ParseForm()
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	text := req.Form.Get("text")
	text = strings.TrimSpace(text)
	tokens := strings.Fields(text)
	var selection string
	if len(tokens) == 0 {
		selection = db.pickFood()
	} else {
		switch tokens[0] {
		case "-add":
			if len(tokens) < 2 {
				selection = "You have to specify a restaurant!"
			} else {
				selection = db.addFood(tokens[1:])
			}
		case "-delete":
			if len(tokens) < 2 {
				selection = "You have to specify a restaurant!"
			} else {
				selection = db.deleteFood(tokens[1:])
			}
		case "-list":
			selection = db.listFood()
		default:
			selection = db.pickFoodFromCategory(tokens[0])
		}
	}

	rawJSON, err = json.Marshal(db)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = ioutil.WriteFile("/fooddb/food.json", rawJSON, 0666)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := make(map[string]string)
	response["response_type"] = "in_channel"
	response["text"] = selection
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(response)
}

func (db fooddb) list() []string {
	list := make([]string, 0, len(db))
	for restaurant := range db {
		list = append(list, restaurant)
	}
	return list
}

func (db fooddb) invert() map[string][]string {
	inverted := make(map[string][]string)
	for key, values := range db {
		for _, value := range values {
			inverted[value] = append(inverted[value], key)
		}
	}
	return inverted
}

func (db fooddb) pickFood() string {
	list := db.list()
	if len(list) == 0 {
		return "No restaurants found!"
	}
	rand.Seed(time.Now().Unix())
	return "You should go to " + list[rand.Intn(len(list))]
}

func (db fooddb) addFood(args []string) string {
	if args[0][0] == '-' {
		if len(args) == 1 {
			return "You have to specify a restaurant!"
		}
		if len(args[0]) == 1 {
			return "You can't just put a dash and not follow it with a category"
		}
		restaurant := strings.Join(args[1:], " ")
		category := args[0][1:]
		db[restaurant] = append(db[restaurant], category)
		return "Successfully added " + restaurant + " to the " + category + " category."
	}
	restaurant := strings.Join(args, " ")
	if _, isPresent := db[restaurant]; !isPresent {
		db[restaurant] = []string{}
	}
	return "Successfully added " + restaurant
}

func (db fooddb) deleteFood(args []string) string {
	restaurant := strings.Join(args, " ")
	delete(db, restaurant)
	return "Successfully deleted " + restaurant
}

func (db fooddb) listFood() string {
	var lines []string
	for restaurant, tags := range db {
		line := restaurant + ": " + strings.Join(tags, ", ")
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (db fooddb) pickFoodFromCategory(category string) string {
	inverted := db.invert()
	if _, isPresent := inverted[category]; !isPresent {
		return "Couldn't find any restaurants in the " + category + " category!"
	}
	rand.Seed(time.Now().Unix())
	return inverted[category][rand.Intn(len(inverted[category]))]
}

// RegisterFoodRoutes registers the routes for the food service
func (api *API) RegisterFoodRoutes(router *mux.Router) {
	router.HandleFunc("/food", api.FoodSuggestion).Methods("GET")
}
