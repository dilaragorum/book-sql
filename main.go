package main

import (
	"database/sql"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

// Golang <----DRIVER(pq)-----> DB PostgreSQL

type Book struct {
	isbn   string
	title  string
	author string
	price  float32
}

// *sql.DB türünde connectionPool tanımlıyorum. Global tanımlamamın sebebi aşağıdaki örnekte init içerisinde
// connectionPool'a atama yapıp, getBooks metodunda da bu connection'ını kitaplar DB'den getirmek için kullanıyorum.
// We use the init() function to set up our connection pool and assign it to the global variable db.
//We're using a global variable to store the connection pool
//because it's an easy way of making it available to our HTTP handlers – but it's by no means the only way.
var connectionPool *sql.DB

func init() {
	var err error
	connectionPool, err = sql.Open("postgres", "postgres://postgres:postgres@localhost?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	router := httprouter.New()
	router.GET("/books", GetBooks)
	router.GET("/books/:isbn", GetBook)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func GetBook(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	isbn := ps.ByName("isbn")
	if isbn == "" {
		http.Error(w, "ISBN cannot be empty", http.StatusBadRequest)
		return
	}

	book := Book{}
	//DB.QueryRow() is used for SELECT queries which return a single row.
	row := connectionPool.QueryRow("SELECT * FROM books WHERE isbn = $1", isbn)
	err := row.Scan(&book.isbn, &book.title, &book.author, &book.price)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "error when scanning book struct", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s %s %s %f\n", book.isbn, book.title, book.author, book.price)
}

func GetBooks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	books, err := GetAllBooksFromDB(connectionPool)
	if err != nil {
		http.Error(w, "Error when getting all books from db", http.StatusInternalServerError)
		return
	}

	for _, book := range books {
		fmt.Fprintf(w, "%s %s %s %f\n", book.isbn, book.title, book.author, book.price)
	}
}

// Kitapları DB'den getirebilmek için yukarıdaki connection'u burada kullanmamız gerekiyor Query yaparken.
func GetAllBooksFromDB(connectionPool *sql.DB) ([]Book, error) {
	//DB.Query() is used for SELECT queries which return multiple rows.
	rows, err := connectionPool.Query("SELECT * FROM books")
	if err != nil {
		return []Book{}, err
	}
	defer rows.Close() // açmış olduğum connection'ını işim bitince pool'a geri atıyorum.

	books := make([]Book, 0)
	for rows.Next() {
		bk := Book{}
		err := rows.Scan(&bk.isbn, &bk.title, &bk.author, &bk.price)
		if err != nil {
			return []Book{}, err
		}
		books = append(books, bk)
	}
	err = rows.Err()
	if err != nil {
		return []Book{}, err
	}

	return books, nil
}
