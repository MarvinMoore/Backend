package main

import (
	"backend/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

type jsonResp struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (app *application) getOneMovie(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.logger.Print(errors.New("invalid id parameter"))
		app.errorJSON(w, err)
		return
	}

	movie, err := app.models.DB.Get(id)

	err = app.writeJSON(w, http.StatusOK, movie, "movie")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}
func (app *application) getAllDiscounts(w http.ResponseWriter, r *http.Request) {
	discounts, err := app.models.DB.GetDiscounts()
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	bytes, _ := json.Marshal(discounts)
	w.Write(bytes)
}
func (app *application) getDiscount(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id") //id="c1"
	intVar, err := strconv.Atoi(id)
	discount, err := app.models.DB.NewGetDiscount(intVar)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]interface{})
	resp["id"] = discount[0].ID
	resp["name"] = discount[0].Name
	resp["discount"] = discount[0].Discount
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}
func (app *application) updateShipping(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered updatesshipping")
	query := r.URL.Query()
	id := query.Get("id")
	intVar, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
	}
	app.models.DB.FlipShipping(intVar)
	w.WriteHeader(200)
}
func (app *application) getNewItem(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id") //id="c1"
	intVar, err := strconv.Atoi(id)
	/*if err != nil {
		fmt.Println(err)
		app.logger.Print(errors.New("invalid id parameter"))
		app.errorJSON(w, err)
		return
	}*/

	item, err := app.models.DB.NewGetItem(intVar)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]interface{})
	resp["id"] = item[0].ID
	resp["topic"] = item[0].Topic
	resp["item_name"] = item[0].ItemName
	resp["description"] = item[0].Description
	resp["image"] = item[0].Image
	resp["quantity"] = item[0].Quantity
	resp["price"] = item[0].Price
	resp["created"] = item[0].Created
	resp["updated"] = item[0].Updated
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	/*err = app.writeJSON(w, http.StatusOK, item, nil)
	if err != nil {
		app.errorJSON(w, err)
		return
	}*/
}

func (app *application) getAllMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := app.models.DB.All()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, movies, "movies")
	if err != nil {
		app.errorJSON(w, err)
		return
	}

}

func (app *application) getAllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := app.models.DB.GenresAll()
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, genres, "genres")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

type Item struct {
	ItemName        string `json:"item_name"`
	ItemPrice       string `json:"item_price"`
	ItemDescription string `json:"item_description"`
	ItemURL         string `json:"item_url"`
}

func (app *application) deleteItem(w http.ResponseWriter, r *http.Request) {
	var itemData Item

	err := json.NewDecoder(r.Body).Decode(&itemData)
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized"))
		return
	}

	err = app.writeJSON(w, http.StatusOK, itemData, "deleted-item")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	fmt.Printf("DELETED ITEM!\n", itemData)
}

type SearchNameStruct struct {
	ItemName string `json:"item_name", db:"item_name"`
}

func (app *application) getItem(w http.ResponseWriter, r *http.Request) {
	search := &SearchNameStruct{}
	err := json.NewDecoder(r.Body).Decode(&search)
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized"))
		return
	}
	row, err := app.models.DB.QuerySearchItem(search.ItemName)
	fmt.Println(row, err)
	if err == nil {
		fmt.Println("Item Exists in DB!", row)
		err = app.writeJSON(w, http.StatusOK, row, "response")
		if err != nil {
			app.errorJSON(w, err)
			return
		}
	} else {
		fmt.Println("Item does not exist!")
		err = app.writeJSON(w, http.StatusForbidden, nil, "Item does not exist.")
		if err != nil {
			app.errorJSON(w, err)
			return
		}
	}
}

func (app *application) getItems(w http.ResponseWriter, r *http.Request) {
	allItems, err := app.models.DB.ItemsALL()
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	size := len(allItems)
	err = app.writeJSON(w, http.StatusOK, allItems, "items")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	for i := 0; i < size; i++ {
		fmt.Println("PASSED ITEMS!\n", *allItems[i])
	}
}

func (app *application) getAllMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	genreID, err := strconv.Atoi(params.ByName("genre_id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	movies, err := app.models.DB.All(genreID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, movies, "movies")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

func (app *application) deleteMovie(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.models.DB.DeleteMovie(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	ok := jsonResp{
		OK: true,
	}

	err = app.writeJSON(w, http.StatusOK, ok, "response")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

func (app *application) insertMovie(w http.ResponseWriter, r *http.Request) {

}

type MoviePayload struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Year        string `json:"year"`
	ReleaseDate string `json:"release_date"`
	Runtime     string `json:"runtime"`
	Rating      string `json:"rating"`
	MPAARating  string `json:"mpaa_rating"`
}

func (app *application) editMovie(w http.ResponseWriter, r *http.Request) {
	var payload MoviePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}

	var movie models.Movie

	if payload.ID != "0" {
		id, _ := strconv.Atoi(payload.ID)
		m, _ := app.models.DB.Get(id)
		movie = *m
		movie.UpdatedAt = time.Now()
	}

	movie.ID, _ = strconv.Atoi(payload.ID)
	movie.Title = payload.Title
	movie.Description = payload.Description
	movie.ReleaseDate, _ = time.Parse("2006-01-02", payload.ReleaseDate)
	movie.Year = movie.ReleaseDate.Year()
	movie.Runtime, _ = strconv.Atoi(payload.Runtime)
	movie.Rating, _ = strconv.Atoi(payload.Rating)
	movie.MPAARating = payload.MPAARating
	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	if movie.ID == 0 {
		err = app.models.DB.InsertMovie(movie)
		if err != nil {
			app.errorJSON(w, err)
			return
		}
	} else {
		err = app.models.DB.UpdateMovie(movie)
		if err != nil {
			app.errorJSON(w, err)
			return
		}
	}

	ok := jsonResp{
		OK: true,
	}

	err = app.writeJSON(w, http.StatusOK, ok, "response")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

func (app *application) searchMovies(w http.ResponseWriter, r *http.Request) {

}
