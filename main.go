package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

// Golang <----DRIVER(pq)-----> DB PostgreSQL

type Book struct {
	Isbn   string  `json:"isbn"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float32 `json:"price"`
}

// *sql.DB türünde connectionPool tanımlıyorum. Global tanımlamamın sebebi aşağıdaki örnekte init içerisinde
// connectionPool'a atama yapıp, getBooks metodunda da bu connection'ını kitaplar DB'den getirmek için kullanıyorum.
// We use the init() function to set up our connection pool and assign it to the global variable db.
//We're using a global variable to store the connection pool
//because it's an easy way of making it available to our HTTP handlers – but it's by no means the only way.
var connectionPool *sql.DB

func init() {
	var err error
	// Go driver aracılığı ile Postgresql veritabanına bağlanıp connectionPool döndürüyor (mesela 100 connection).
	connectionPool, err = sql.Open("postgres", "postgres://postgres:postgres@localhost?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	router := httprouter.New()
	router.GET("/books", GetBooks)
	router.GET("/books/:isbn", GetBook)
	router.POST("/books/create", CreateBook)

	log.Fatal(http.ListenAndServe(":8080", router))
}

/*
{
	isbn: "",
 	author: "",
	title: "",
	price: 10
}
*/

/*
curl -X POST localhost:8080/books/create \
-H 'Content-Type: application/json' \
-d '{ "isbn": "dilara", "author": "Dilara", "title": "Dilaranın Kitabı", "price": 0 }'
*/
func CreateBook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var book Book
	// json.NewDecoder -  istek atılan json request body'i go struct'ımıza decode ediyoruz.
	// Sonra decode ettiğimiz json'u, Decode(&book) ile book'a atıyoruz.
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		http.Error(w, "body malformed", http.StatusBadRequest)
		return
	}

	// DB.Exec() is used for statements which don’t return rows (like INSERT and DELETE).
	result, err := connectionPool.
		Exec("INSERT INTO books VALUES($1, $2, $3, $4)", book.Isbn, book.Title, book.Author, book.Price)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// The sql.Result() interface guarantees two methods: LastInsertId() – which is often used to return
	//the value of an new auto increment id, and RowsAffected() – which contains the number of rows
	//that the statement affected.
	numberOfRowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Book %s created successfully (%d row affected)\n", book.Isbn, numberOfRowsAffected)
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
	err := row.Scan(&book.Isbn, &book.Title, &book.Author, &book.Price)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "error when scanning book struct", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s %s %s %f\n", book.Isbn, book.Title, book.Author, book.Price)
}

func GetBooks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	books, err := GetAllBooksFromDB(connectionPool)
	if err != nil {
		http.Error(w, "Error when getting all books from db", http.StatusInternalServerError)
		return
	}

	for _, book := range books {
		fmt.Fprintf(w, "%s %s %s %f\n", book.Isbn, book.Title, book.Author, book.Price)
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
		err := rows.Scan(&bk.Isbn, &bk.Title, &bk.Author, &bk.Price)
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
