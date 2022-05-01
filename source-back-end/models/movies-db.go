package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DBModel struct {
	DB *sql.DB
}

func GetDB() *sql.DB {
	mainDB, err := sql.Open("mysql",
		"GolangAdmin:MarvinRulez!@tcp(35.184.228.229:3306)/shopdata")
	if err != nil {
		log.Fatal(err)
	}
	//defer mainDB.Close()
	return mainDB
}
func (m *DBModel) GetDiscounts() ([]*NewDiscountList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	mainDB := GetDB()
	query := `select id, name, discount from base_discountcode`
	rows, err := mainDB.QueryContext(ctx, query)
	if err != nil {
		fmt.Println("ERR:", err)
		return nil, err
	}
	defer rows.Close()
	var discounts []*NewDiscountList

	for rows.Next() {
		var i NewDiscountList
		err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Discount,
		)
		if err != nil {
			fmt.Println("SCAN ERR:", err)
			return nil, err
		}
		discounts = append(discounts, &i)
	}
	fmt.Println("Returning: ", discounts)
	return discounts, nil
}
func (m *DBModel) NewGetDiscount(id int) ([]*NewDiscountList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	mainDB := GetDB()
	query := `select id, name, discount from base_discountcode where id = ?`
	rows, err := mainDB.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println("ERR:", err)
		return nil, err
	}
	defer rows.Close()
	var discount []*NewDiscountList
	for rows.Next() {
		var i NewDiscountList
		err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Discount,
		)
		if err != nil {
			fmt.Println("SCAN ERR:", err)
			return nil, err
		}
		discount = append(discount, &i)
	}
	fmt.Println("Returning: ", discount)
	return discount, nil
}

type Order struct {
	ID            int
	PaymentMethod string
	TaxPrice      float64
	ShippingPrice float64
	Totalprice    float64
	IsPaid        int
	Created       []uint8
	UserID        int
	isShipped     int
}

func (m *DBModel) FlipShipping(id int) error {
	mainDB := GetDB()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `select id, paymentmethod, taxprice, shippingprice, totalprice, ispaid, created, user_id, isshipped from base_order where id = ?`
	rows, err := mainDB.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println("ERR:", err)
		return err
	}
	var orders []*Order
	for rows.Next() {
		var i Order
		err := rows.Scan(
			&i.ID,
			&i.PaymentMethod,
			&i.TaxPrice,
			&i.ShippingPrice,
			&i.Totalprice,
			&i.IsPaid,
			&i.Created,
			&i.UserID,
			&i.isShipped,
		)
		if err != nil {
			fmt.Println("SCAN ERR:", err)
			i.Created = []uint8{2, 3}
			//return nil, err
		}
		orders = append(orders, &i)
	}
	fmt.Println(orders[0].ID)
	fmt.Println(orders[0].isShipped)
	if orders[0].isShipped == 0 {
		fmt.Println("flipping to no")
		if _, err := mainDB.Query("UPDATE base_order SET isshipped = ? WHERE id = ?", 1, id); err != nil {
			// If there is any issue with inserting into the database, return a 500 error
			fmt.Println(err)
			return err
		}
	}
	if orders[0].isShipped == 1 {
		fmt.Println("flipping to yes")
		if _, err := mainDB.Query("update base_order set isshipped = 0 where id = ?", id); err != nil {
			// If there is any issue with inserting into the database, return a 500 error
			fmt.Println(err)
			return err
		}

	}
	fmt.Println("Successfully flipped shipping")
	return nil

}
func (m *DBModel) NewGetItem(id int) ([]*NewItemList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	//connect to DB
	mainDB := GetDB()
	query := `select id, item_name, description, image, quantity, price, topic from base_item where id = ?`
	rows, err := mainDB.QueryContext(ctx, query, id)
	if err != nil {
		fmt.Println("ERR:", err)
		return nil, err
	}
	defer rows.Close()

	var items []*NewItemList

	for rows.Next() {
		var i NewItemList
		err := rows.Scan(
			&i.ID,
			&i.ItemName,
			&i.Description,
			&i.Image,
			&i.Quantity,
			&i.Price,
			//&i.Updated,
			//&i.Created,
			&i.Topic,
		)
		if err != nil {
			fmt.Println("SCAN ERR:", err)
			i.Topic = "NULL"
			//return nil, err
		}
		i.Image = "http://35.224.232.15/media/" + i.Image
		items = append(items, &i)
	}
	fmt.Println("Returning: ", items)
	return items, nil
}

// Get returns one movie and error, if any
func (m *DBModel) Get(id int) (*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, title, description, year, release_date, rating, runtime, mpaa_rating,
				created_at, updated_at from movies where id = $1
	`

	row := m.DB.QueryRowContext(ctx, query, id)

	var movie Movie

	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.Year,
		&movie.ReleaseDate,
		&movie.Rating,
		&movie.Runtime,
		&movie.MPAARating,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// get genres, if any
	query = `select
				mg.id, mg.movie_id, mg.genre_id, g.genre_name
			from
				movies_genres mg
				left join genres g on (g.id = mg.genre_id)
			where
				mg.movie_id = $1
	`

	rows, _ := m.DB.QueryContext(ctx, query, id)
	defer rows.Close()

	genres := make(map[int]string)
	for rows.Next() {
		var mg MovieGenre
		err := rows.Scan(
			&mg.ID,
			&mg.MovieID,
			&mg.GenreID,
			&mg.Genre.GenreName,
		)
		if err != nil {
			return nil, err
		}
		genres[mg.ID] = mg.Genre.GenreName
	}

	movie.MovieGenre = genres

	return &movie, nil
}

// All returns all movies and error, if any
func (m *DBModel) All(genre ...int) ([]*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	where := ""
	if len(genre) > 0 {
		where = fmt.Sprintf("where id in (select movie_id from movies_genres where genre_id = %d)", genre[0])
	}

	query := fmt.Sprintf(`select id, title, description, year, release_date, rating, runtime, mpaa_rating,
				created_at, updated_at from movies  %s order by title`, where)

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*Movie

	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Description,
			&movie.Year,
			&movie.ReleaseDate,
			&movie.Rating,
			&movie.Runtime,
			&movie.MPAARating,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// get genres, if any
		genreQuery := `select
			mg.id, mg.movie_id, mg.genre_id, g.genre_name
		from
			movies_genres mg
			left join genres g on (g.id = mg.genre_id)
		where
			mg.movie_id = $1
		`

		genreRows, _ := m.DB.QueryContext(ctx, genreQuery, movie.ID)

		genres := make(map[int]string)
		for genreRows.Next() {
			var mg MovieGenre
			err := genreRows.Scan(
				&mg.ID,
				&mg.MovieID,
				&mg.GenreID,
				&mg.Genre.GenreName,
			)
			if err != nil {
				return nil, err
			}
			genres[mg.ID] = mg.Genre.GenreName
		}
		genreRows.Close()

		movie.MovieGenre = genres
		movies = append(movies, &movie)

	}
	return movies, nil
}

func (m *DBModel) GenresAll() ([]*Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, genre_name, created_at, updated_at from genres order by genre_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*Genre

	for rows.Next() {
		var g Genre
		err := rows.Scan(
			&g.ID,
			&g.GenreName,
			&g.CreatedAt,
			&g.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		genres = append(genres, &g)
	}

	return genres, nil
}
func (m *DBModel) QueryUser(user string, hashedPassword string) error {
	isAdmin := "no"
	fmt.Println("Entered query user with", user, hashedPassword)
	if _, err := m.DB.Query("insert into users_new values ($1, $2, $3)", user, hashedPassword, isAdmin); err != nil {
		// If there is any issue with inserting into the database, return a 500 error
		return err
	}
	return nil
}
func (m *DBModel) QuerySelectPass(user string) *sql.Row {
	fmt.Println("Entered select pass with", user)
	result := m.DB.QueryRow("select password from users_new where username=$1", user)
	return result
}
func (m *DBModel) QueryAdmin(user string) bool {
	result := m.DB.QueryRow("select isAdmin from users_new where username=$1", user)
	fmt.Println(result)
	var i UserNew
	err1 := result.Scan(
		&i.IsAdmin,
	)
	if err1 != nil {
		fmt.Println(err1)
	}
	fmt.Println(i.IsAdmin)
	if i.IsAdmin == "yes" {
		return true
	} else {
		return false
	}
}
func (m *DBModel) QuerySearchItem(query string) ([]*ItemList, error) {
	sqlStmt := `SELECT item_name FROM item_table WHERE item_name = $1`
	err := m.DB.QueryRow(sqlStmt, query).Scan(&query)
	rows := m.DB.QueryRow("select item_name, item_price, item_description, item_url from item_table where item_name=$1", query)
	var items []*ItemList

	var i ItemList
	err1 := rows.Scan(
		&i.ItemName,
		&i.ItemPrice,
		&i.ItemDescription,
		&i.ItemURL,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if err1 != nil {
		log.Println(err1)
		return nil, err1
	}
	items = append(items, &i)

	fmt.Println("Passing Result!", items)
	return items, nil
}
func (m *DBModel) InsertMovie(movie Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into movies (title, description, year, release_date, runtime, rating, mpaa_rating,
				created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := m.DB.ExecContext(ctx, stmt,
		movie.Title,
		movie.Description,
		movie.Year,
		movie.ReleaseDate,
		movie.Runtime,
		movie.Rating,
		movie.MPAARating,
		movie.CreatedAt,
		movie.UpdatedAt,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
func (m *DBModel) UpdateAdminDB(user string, action string) error {
	fmt.Println("Entered Update AdminDB")
	sqlStatement := `UPDATE users_new SET isadmin = $2 WHERE username = $1`
	_, err := m.DB.Exec(sqlStatement, user, action)
	if err != nil {
		log.Println(err)
		return err
	}
	fmt.Println("Successfuly updated: ")
	return nil

}
func (m *DBModel) UpdateMovie(movie Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update movies set title = $1, description = $2, year = $3, release_date = $4, 
				runtime = $5, rating = $6, mpaa_rating = $7,
				updated_at = $8 where id = $9`

	_, err := m.DB.ExecContext(ctx, stmt,
		movie.Title,
		movie.Description,
		movie.Year,
		movie.ReleaseDate,
		movie.Runtime,
		movie.Rating,
		movie.MPAARating,
		movie.UpdatedAt,
		movie.ID,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m *DBModel) DeleteMovie(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "delete from movies where id = $1"

	_, err := m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) ItemsALL() ([]*ItemList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select item_name, item_price, item_description, item_url from item_table order by item_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*ItemList

	for rows.Next() {
		var i ItemList
		err := rows.Scan(
			&i.ItemName,
			&i.ItemPrice,
			&i.ItemDescription,
			&i.ItemURL,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &i)
	}

	return items, nil
}

func (m *DBModel) InsertItem(item ItemList) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into item_table (item_name, item_price, item_description, item_url) values ($1, $2, $3, $4)`

	_, err := m.DB.ExecContext(ctx, stmt,
		item.ItemName,
		item.ItemPrice,
		item.ItemDescription,
		item.ItemURL,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
