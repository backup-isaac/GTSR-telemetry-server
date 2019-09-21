package api

/*
This API represents the /food Slack command, that tells you where you should eat.
Slack users can manage this database for all of solar using certain commands.
You can also choose to filter out places that are closed if you're looking
to eat during Real Solar Hours^TM
*/

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// FoodHandler handles requests related to the food bot Slack plugin
type FoodHandler struct{}

// NewFoodHandler is the basic FoodHandler constructor
func NewFoodHandler() *FoodHandler {
	return &FoodHandler{}
}

const helpText = `Welcome to the food bot! Type /food to get a food suggestion.

If you're in the mood for something specific, type /food followed by a tag:
	/food fast

To add a restaurant to the database, type:
	/food -add Cookout

To give a restaurant a category:
	/food -add -fast Cookout
Now Cookout has the fast tag. Tags can be whatever you want.

To specify when a restaurant closes, type:
	/food -add -closes=4am Cookout
Other options for the closes tag syntax:
	/food -add -closes=9:30pm Wagaya
	/food -add -closes=never Waffle House
Any other syntax for the closes tag will get treated as a normal tag

If it's real solar hours, you can filter out closed restaurants by using:
	/food -late

To delete a restaurant, type:
	/food -delete Krystal

To list all the restaurants in the database, type:
	/food -list

Happy eating!`

var foodLock sync.Mutex

type fooddb map[string][]string

const closesTagRegex = "closes=((((([1-9])|(1[012]))|(([1-9])|(1[012])):[0-5][0-9])((am)|(pm)))|never)"

// FoodSuggestion handles a request from Slack asking for a suggestion of
// where to eat
func (f *FoodHandler) FoodSuggestion(res http.ResponseWriter, req *http.Request) {
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
	if strings.Contains(text, "\"") {
		selection = "No double quotes are allowed in the text string"
	} else if len(tokens) == 0 {
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
		case "-late":
			if len(tokens) == 1 {
				selection = db.filterOpen().pickFood()
			} else {
				selection = db.filterOpen().pickFoodFromCategory(tokens[1])
			}
		case "-help":
			selection = helpText
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
		if isCloseTag, _ := regexp.MatchString(closesTagRegex, category); isCloseTag {
			// Remove other close tags if this one is a close tag
			var newtags []string
			for _, otherCategory := range db[restaurant] {
				if isAlsoCloseTag, _ := regexp.MatchString(closesTagRegex, otherCategory); !isAlsoCloseTag {
					newtags = append(newtags, otherCategory)
				}
			}
			db[restaurant] = newtags
		} else {
			if strings.HasPrefix(category, "closes=") {
				return "closes= tag has invalid syntax!"
			}
		}
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
	return "You should go to " + inverted[category][rand.Intn(len(inverted[category]))]
}

func (db fooddb) filterOpen() fooddb {
	var newdb fooddb = make(map[string][]string)
	loc, _ := time.LoadLocation("America/New_York")
	curtime := time.Now().In(loc)
	for restaurant, tags := range db {
		for _, tag := range tags {
			match, err := regexp.MatchString(closesTagRegex, tag)
			if err != nil {
				log.Printf("Error parsing time tag regex: %s\n", err)
				return newdb
			}
			if match {
				timestr := tag[len("closes="):]
				if timestr == "never" {
					newdb[restaurant] = tags
					continue
				}
				var hourstr string
				var minutestr string
				if colonIndex := strings.Index(timestr, ":"); colonIndex >= 0 {
					hourstr = timestr[:colonIndex]
					minutestr = timestr[colonIndex+1 : len(timestr)-2]
				} else {
					hourstr = timestr[:len(timestr)-2]
					minutestr = "00"
				}
				hour, err := strconv.Atoi(hourstr)
				if err != nil {
					log.Printf("Error parsing time tag hour: %s\n", err)
					return newdb
				}
				minute, err := strconv.Atoi(minutestr)
				if err != nil {
					log.Printf("Error parsing time tag minute: %s\n", err)
				}
				if timestr[len(timestr)-2:] == "pm" && hour != 12 {
					hour += 12
				} else if timestr[len(timestr)-2:] == "am" && hour == 12 {
					hour -= 12
				}
				if (curtime.Hour() < 6 && hour < 6 && curtime.Hour() < hour) ||
					(curtime.Hour() > 6 && hour < 6) ||
					(curtime.Hour() > 6 && curtime.Hour() < hour) ||
					(curtime.Hour() == hour && curtime.Minute() < minute) {
					newdb[restaurant] = tags
				}
			}
		}
	}
	return newdb
}

// RegisterRoutes registers the routes for the food service
func (f *FoodHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/food", f.FoodSuggestion).Methods("GET")
}
