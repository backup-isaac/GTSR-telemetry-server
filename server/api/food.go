package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// FoodSuggestion handles a request from Slack asking for a suggestion of
// where to eat
func (api *API) FoodSuggestion(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	text := req.Form.Get("text")
	text = strings.TrimSpace(text)
	tokens := strings.Fields(text)
	var selection string
	if len(tokens) == 0 {
		selection, err = pickFood()
		if err != nil {
			log.Printf("%s\n", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		switch tokens[0] {
		case "-add":
			if len(tokens) < 2 {
				selection = "You have to specify a restaurant!"
			} else {
				selection = addFood(tokens[1:])
			}
		case "-delete":
			if len(tokens) < 2 {
				selection = "You have to specify a restaurant!"
			} else {
				selection = deleteFood(tokens[1:])
			}
		case "-list":
			selection = listFood()
		default:
			selection, err = pickFoodFromCategory(tokens[0])
			if err != nil {
				log.Printf("Error getting food from category: %s", err)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	response := make(map[string]string)
	response["response_type"] = "in_channel"
	response["text"] = selection
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(response)
}

func addFood(args []string) string {
	if args[0][0] == '-' {
		if len(args) == 1 {
			return "You have to specify a restaurant!"
		}
		if len(args[0]) == 1 {
			return "You can't just put a dash and not follow it with a category"
		}
		restaurant := strings.Join(args[1:], " ")
		err := addRestaurantTo("/fooddb/food_"+args[0][1:]+".txt", restaurant)
		if err != nil {
			return "Failed to add restaurant to database. Try a different category name?"
		}
		return "Successfully added " + restaurant + " to category " + args[0][1:] + "."
	}
	restaurant := strings.Join(args, " ")
	err := addRestaurantTo("/fooddb/other.txt", restaurant)
	if err != nil {
		return "Failed to add restaurant to database"
	}
	return "Successfully added " + restaurant + " to database."
}

func deleteFood(args []string) string {
	restaurant := strings.Join(args, " ")
	files, err := ioutil.ReadDir("/fooddb")
	if err != nil {
		return "Error: couldn't find food database!"
	}
	for _, f := range files {
		fn := path.Join("/fooddb", f.Name())
		category, err := getRestaurantsFromFile(fn)
		if err != nil {
			return "Failed to remove " + restaurant
		}
		var newRestaurantList []string
		for _, newRestaurant := range category {
			if newRestaurant != restaurant {
				newRestaurantList = append(newRestaurantList, newRestaurant)
			}
		}
		if len(newRestaurantList) == 0 {
			err = os.Remove(fn)
			if err != nil {
				return "Failed to remove " + restaurant
			}
		} else {
			ioutil.WriteFile(fn, []byte(strings.Join(newRestaurantList, "\n")), 0666)
		}
	}
	return "Successfully deleted " + restaurant
}

func listFood() string {
	files, err := ioutil.ReadDir("/fooddb")
	if err != nil {
		return "Error: couldn't find food database!"
	}
	restaurantTags := make(map[string][]string)
	for _, f := range files {
		fn := path.Join("/fooddb", f.Name())
		restaurants, err := getRestaurantsFromFile(fn)
		if err != nil {
			return "Failed to list"
		}
		if strings.HasPrefix(f.Name(), "food_") {
			category := f.Name()[len("food_") : len(f.Name())-len(".txt")]
			for _, restaurant := range restaurants {
				restaurantTags[restaurant] = append(restaurantTags[restaurant], category)
			}
		} else {
			for _, restaurant := range restaurants {
				if _, isPresent := restaurantTags[restaurant]; !isPresent {
					restaurantTags[restaurant] = []string{}
				}
			}
		}
	}
	var lines []string
	for restaurant, tags := range restaurantTags {
		line := restaurant + ": " + strings.Join(tags, ", ")
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func addRestaurantTo(fl string, restaurant string) error {
	var restaurants []string
	if _, err := os.Stat(fl); !os.IsNotExist(err) {
		restaurants, err = getRestaurantsFromFile(fl)
		if err != nil {
			return err
		}
	}
	restaurants = removeDuplicates(append(restaurants, restaurant))
	return ioutil.WriteFile(fl, []byte(strings.Join(restaurants, "\n")), 0666)
}

func pickFood() (string, error) {
	files, err := ioutil.ReadDir("/fooddb")
	if err != nil {
		return "", err
	}
	var restaurants []string
	for _, f := range files {
		category, err := getRestaurantsFromFile(path.Join("/fooddb", f.Name()))
		if err != nil {
			return "", err
		}
		for _, restaurant := range category {
			restaurants = append(restaurants, restaurant)
		}
	}
	restaurants = removeDuplicates(restaurants)
	return pickRestaurant(restaurants), nil
}

func pickFoodFromCategory(category string) (string, error) {
	path := "/fooddb/food_" + category + ".txt"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "There aren't any restaurants in that category!", nil
	}
	restaurants, err := getRestaurantsFromFile(path)
	if err != nil {
		return "", err
	}
	return pickRestaurant(restaurants), nil
}

func getRestaurantsFromFile(filename string) ([]string, error) {
	categoryBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	categoryText := strings.TrimSpace(string(categoryBytes))
	return strings.Split(string(categoryText), "\n"), nil
}

func pickRestaurant(restaurants []string) string {
	if len(restaurants) == 0 {
		return "There aren't any restaurants to choose from!"
	}
	rand.Seed(time.Now().Unix())
	return "You should go to " + restaurants[rand.Intn(len(restaurants))]
}

func removeDuplicates(items []string) []string {
	seen := make(map[string]struct{})
	var ret []string
	for _, item := range items {
		if _, isSeen := seen[item]; !isSeen {
			ret = append(ret, item)
			seen[item] = struct{}{}
		}
	}
	return ret
}

// RegisterFoodRoutes registers the routes for the food service
func (api *API) RegisterFoodRoutes(router *mux.Router) {
	router.HandleFunc("/food", api.FoodSuggestion).Methods("GET")
}
