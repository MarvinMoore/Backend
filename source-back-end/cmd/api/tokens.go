package main

import (
	"backend/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pascaldekloe/jwt"
	"golang.org/x/crypto/bcrypt"
)

var validUser = models.User{
	ID:       10,
	Email:    "me@here.com",
	Password: "$2a$12$MAYkOm4QWArsV5kczREsT.u7TZjyO7srH1lY6CI29WTSSe3DHJaMG",
	ISAdmin:  0,
}

type Credentials struct {
	Password string `json:"password", db:"password"`
	Username string `json:"username", db:"username"`
}
type AdminStruct struct {
	Username string `json:"username"`
	IsAdmin  string `json:"isAdmin"`
}

type DBModel struct {
	DB *sql.DB
}

// this map stores the users sessions. For larger scale applications, you can use a database or cache for this purpose
var sessions = map[string]session{}

// each session contains the username of the user and the time at which it expires
type session struct {
	username string
	expiry   time.Time
}

// we'll use this method later to determine if the session has expired
func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

func (app *application) Signup(w http.ResponseWriter, r *http.Request) {

	// Parse and decode the request body into a new `Credentials` instance
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Salt and hash the password using the bcrypt algorithm
	// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	err = app.models.DB.QueryUser(creds.Username, string(hashedPassword))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("Couldn't add user/pass in DB!")
		fmt.Println(err)
	}
	// Next, insert the username, along with the hashed password into the database

	// We reach this point if the credentials we correctly stored in the database, and the default status of 200 is sent back
	app.writeJSON(w, http.StatusOK, creds, "response")
}

func (app *application) Signin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered Sign in!")
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized"))
		return
	}

	result := app.models.DB.QuerySelectPass(creds.Username)
	if err != nil {
		// If there is an issue with the database, return a 500 error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// We create another instance of `Credentials` to store the credentials we get from the database
	storedCreds := &Credentials{}
	err = result.Scan(&storedCreds.Password)
	if err != nil {
		fmt.Println(err)
		// If an entry with the username does not exist, send an "Unauthorized"(401) status
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// If the error is of any other type, send a 500 status
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Compare the stored hashed password, with the hashed version of the password that was received
	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		// If the two passwords don't match, return a 401 status
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
	}
	// If we reach this point, that means the users password was correct, and that they are authorized
	// The default 200 status is sent
	var claims jwt.Claims
	type Approval struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	claims.Subject = fmt.Sprint(validUser.ID)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "mydomain.com"
	claims.Audiences = []string{"mydomain.com"}
	claims.Set = map[string]interface{}{
		"approved": []Approval{{"RPG-7", 1}},
	}
	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secret))
	if err != nil {
		app.errorJSON(w, errors.New("error signing"))
		return
	}
	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)
	// Set the token in the session map, along with the session information
	sessions[sessionToken] = session{
		username: creds.Username,
		expiry:   expiresAt,
	}
	// Finally, we set the client cookie for "session_token" as the session token we just generated
	// we also set an expiry time of 120 seconds
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})
	fmt.Println("Finished signin!")
	app.writeJSON(w, http.StatusOK, string(jwtBytes), "reponse")
}
func (app *application) Welcome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered Welcome!")
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println(err)
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value
	fmt.Println(sessionToken)
	// We then get the session from our session map
	userSession, exists := sessions[sessionToken]

	if !exists {
		// If the session token is not present in session map, return an unauthorized error
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// If the session is present, but has expired, we can delete the session, and return
	// an unauthorized status
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// If the session is valid, return the welcome message to the user
	w.Write([]byte(fmt.Sprintf("Welcome %s!", userSession.username)))
}
func (app *application) addItem(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println(err)
		if err == http.ErrNoCookie {
			app.writeJSON(w, http.StatusOK, nil, "Please sign in.")
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value
	fmt.Println(sessionToken)
	userSession, exists := sessions[sessionToken]

	if !exists {
		app.writeJSON(w, http.StatusOK, nil, "Please sign in.")
		// If the session token is not present in session map, return an unauthorized error
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// If the session is present, but has expired, we can delete the session, and return
	// an unauthorized status
	if userSession.isExpired() {
		app.writeJSON(w, http.StatusOK, nil, "Token Expired.")
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	username := userSession.username
	result := app.models.DB.QueryAdmin(username)
	if result == true {
		fmt.Println("Admin can continue")
	} else {
		app.writeJSON(w, http.StatusOK, nil, "You do not have admin access.")
		return
	}
	fmt.Println("passed check")
	var itemData Item

	err1 := json.NewDecoder(r.Body).Decode(&itemData)
	if err1 != nil {
		log.Println(err1)
		app.errorJSON(w, errors.New("unauthorized"))
		return
	}

	err = app.writeJSON(w, http.StatusOK, itemData, "added-item")
	if err != nil {
		log.Println(err)
		app.errorJSON(w, err)
		return
	}
	itemParsed := models.ItemList{
		ItemName:        itemData.ItemName,
		ItemPrice:       itemData.ItemPrice,
		ItemDescription: itemData.ItemDescription,
		ItemURL:         itemData.ItemURL,
	}
	fmt.Printf("Got Item Data,\n itemname - %s\n itemprice - %s\n itemdescription - %s\n itemurl - %s\n", itemData.ItemName, itemData.ItemPrice, itemData.ItemDescription, itemData.ItemURL)
	err = app.models.DB.InsertItem(itemParsed)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	log.Println("added", itemParsed)
}
func (app *application) checkAdmin(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		fmt.Println(err)
		if err == http.ErrNoCookie {
			app.writeJSON(w, http.StatusOK, nil, "Please sign in.")
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value
	fmt.Println(sessionToken)
	userSession, exists := sessions[sessionToken]

	if !exists {
		app.writeJSON(w, http.StatusOK, nil, "Please sign in.")
		// If the session token is not present in session map, return an unauthorized error
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// If the session is present, but has expired, we can delete the session, and return
	// an unauthorized status
	if userSession.isExpired() {
		app.writeJSON(w, http.StatusOK, nil, "Token Expired.")
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	username := userSession.username
	result := app.models.DB.QueryAdmin(username)
	if result == true {
		app.writeJSON(w, http.StatusOK, nil, "You have admin access.")
	} else {
		app.writeJSON(w, http.StatusForbidden, nil, "You do not have admin access.")
	}
}
func (app *application) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filters := query.Get("search") //search="color"
	row, err := app.models.DB.QuerySearchItem(filters)
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
func (app *application) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Entered Update Admin")
	//First send the DB query, if error, it already exist, or user doesnt exist.
	var creds AdminStruct
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(creds)
	if creds.IsAdmin == "yes" {
		err := app.models.DB.UpdateAdminDB(creds.Username, creds.IsAdmin)
		if err != nil {
			log.Println(err)
			app.writeJSON(w, http.StatusForbidden, err, "Error Encountered")
			return
		}
		app.writeJSON(w, http.StatusOK, nil, "Updated Admin Successfully")
		return
	}
	if creds.IsAdmin == "no" {
		err := app.models.DB.UpdateAdminDB(creds.Username, creds.IsAdmin)
		if err != nil {
			log.Println(err)
			app.writeJSON(w, http.StatusForbidden, err, "Error Encountered")
			return
		}
		app.writeJSON(w, http.StatusOK, nil, "Updated Admin Successfully")
		return
	}

}
