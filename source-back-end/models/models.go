package models

import (
	"database/sql"
	"time"
)

// Models is the wrapper for database
type Models struct {
	DB DBModel
}

// NewModels returns models with db pool
func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{DB: db},
	}
}

// Movie is the type for movies
type Movie struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Year        int            `json:"year"`
	ReleaseDate time.Time      `json:"release_date"`
	Runtime     int            `json:"runtime"`
	Rating      int            `json:"rating"`
	MPAARating  string         `json:"mpaa_rating"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	MovieGenre  map[int]string `json:"genres"`
}

// Genre is the type for genre
type Genre struct {
	ID        int       `json:"id"`
	GenreName string    `json:"genre_name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// MovieGenre is the type for movie genre
type MovieGenre struct {
	ID        int       `json:"-"`
	MovieID   int       `json:"-"`
	GenreID   int       `json:"-"`
	Genre     Genre     `json:"genre"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// User is the type for users
type User struct {
	ID       int
	Email    string
	Password string
	ISAdmin  int
}
type UserNew struct {
	Username string
	Password string
	IsAdmin  string
}
type ItemList struct {
	//category
	//size
	//updated
	//created
	ItemName        string `json:"item_name"`
	ItemPrice       string `json:"item_price"`
	ItemDescription string `json:"item_description"`
	//quantity
	ItemURL string `json:"item_url"`
}
type NewItemList struct {
	ID          int       `json:"id"`
	Topic       string    `json:"topic"`
	ItemName    string    `json:"item_name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}
type NewDiscountList struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Discount float64 `json:"discount"`
}
