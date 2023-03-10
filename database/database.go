package database

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/google/uuid"
)

// exported
type Client struct {
	path string
}

// NewClient -
// construct a client
func NewClient(path string) Client {
	return Client{path}
}

type databaseSchema struct {
	Users map[string]User `json:"users"` // key,value = email,user
	Posts map[string]Post `json:"posts"` // key,value = id, post
}

// User -
type User struct {
	CreatedAt time.Time `json:"createdAt"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
}

// Post -
type Post struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UserEmail string    `json:"userEmail"`
	Text      string    `json:"text"`
}

func (c Client) CreatePost(userEmail, text string) (Post, error) {
	// read db, ensure user exists
	db, err := c.readDB()
	if err != nil {
		return Post{}, err
	}
	if _, ok := db.Users[userEmail]; !ok {
		return Post{}, errors.New("user doesn't exist")
	}

	// create new post and add to db
	post := Post{
		ID:        uuid.New().String(),
		CreatedAt: time.Now().UTC(),
		UserEmail: userEmail,
		Text:      text,
	}

	db.Posts[post.ID] = post // add post to Posts
	err = c.updateDB(db)     // save to disk
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

// GetPosts -
// return all posts of a specific user identified by their userEmail
func (c Client) GetPosts(userEmail string) ([]Post, error) {
	// read db, ensure user exists
	db, err := c.readDB()
	if err != nil {
		return []Post{}, err
	}

	allPosts := []Post{}
	for _, post := range db.Posts {
		if post.UserEmail == userEmail {
			allPosts = append(allPosts, post)
		}
	}

	return allPosts, nil
}

// DeletePost -
// delete a single post identified by the id
func (c Client) DeletePost(id string) error {
	// get db from database
	db, err := c.readDB()
	if err != nil {
		return err
	}
	delete(db.Posts, id) // delete post, if doesn't exist noop
	err = c.updateDB(db) // write to disk
	if err != nil {
		return err
	}
	return nil
}

// create new db file (json) at path specified by the client
// empty databaseSchema
// overwrite any previous data in file if existed previously
func (c Client) createDB() error {
	db := databaseSchema{
		Users: make(map[string]User),
		Posts: make(map[string]Post),
	}
	payload, err := json.Marshal(db)
	if err != nil {
		return err
	}
	err = os.WriteFile(c.path, payload, 0666)
	return err
}

// EnsureDB -
// check if db exists already, if good do nothing, otherwise create it using createDB
func (c Client) EnsureDB() error {
	_, err := os.ReadFile(c.path)
	// create new db if doesn't exist
	if err != nil {
		return c.createDB()
	}
	// already exists, do nothing
	return nil
}

// overwrite db file with the data in given databaseSchema
// databaseSchema has JSON tags, can marshal to json format byte slice
func (c Client) updateDB(db databaseSchema) error {
	payload, err := json.Marshal(db)
	if err != nil {
		return err
	}
	err = os.WriteFile(c.path, payload, 0666)
	return err
}

// return data read from db at path in client as a databaseSchema
func (c Client) readDB() (databaseSchema, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return databaseSchema{}, err
	}

	// convert data from json byte slice to databaseSchema
	db := databaseSchema{}
	err = json.Unmarshal(data, &db)
	if err != nil {
		return databaseSchema{}, err
	}

	return db, nil
}

// CreateUser -
// email needs to be unique for each user
func (c Client) CreateUser(email, password, name string, age int) (User, error) {
	// read current status of db
	db, err := c.readDB()
	if err != nil {
		return User{}, err
	}

	// create new user
	newUser := User{
		CreatedAt: time.Now().UTC(),
		Email:     email,
		Password:  password,
		Name:      name,
		Age:       age,
	}

	// add newUser and write to disk
	db.Users[email] = newUser
	err = c.updateDB(db)
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

// UddateUser -
// similar to CreateUser but return an error if user doesn't already exist
// do not update CreatedAt timestamp
func (c Client) UpdateUser(email, password, name string, age int) (User, error) {
	// read from db to see if user already exists
	db, err := c.readDB()
	if err != nil {
		return User{}, err
	}

	// check if email is a key in db.Users
	if _, ok := db.Users[email]; !ok {
		return User{}, errors.New("user doesn't exist")
	}
	// user does exist, we will update (email and CreatedAt fields won't change)
	user := db.Users[email]

	user.Password = password
	user.Name = name
	user.Age = age

	db.Users[email] = user
	err = c.updateDB(db)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// GetUser -
// return user given the email from the db
func (c Client) GetUser(email string) (User, error) {
	db, err := c.readDB()
	if err != nil {
		return User{}, err
	}

	if user, ok := db.Users[email]; ok {
		return user, nil
	} else {
		return User{}, errors.New("user doesn't exist")
	}
}

// DeleteUser -
// delete a user (via email key) from db
func (c Client) DeleteUser(email string) error {
	db, err := c.readDB()
	if err != nil {
		return err
	}

	delete(db.Users, email) // if key doesn't exist, this is no-op
	err = c.updateDB(db)    // save changes to disk
	if err != nil {
		return err
	}

	return nil
}
