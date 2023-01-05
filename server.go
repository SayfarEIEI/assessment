package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Expenses struct {
	Id     int      `json:"id"`
	Title  string   `json:"title"`
	Amount int      `json:"amount"`
	Note   string   `json:"note"`
	Tags   []string `json:"tags"`
}

var expense Expenses

func main() {
	createTable()
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	db, errDb := sql.Open("postgres", os.Getenv("DB_URL"))
	e.GET("/expenses", func(c echo.Context) error {
		var expenses []Expenses
		data, errData := db.Prepare("SELECT id,title,amount,note,tags FROM expenses")
		if errData != nil {
			return c.JSON(http.StatusBadRequest, errData)
		}
		row, err := data.Query()
		if err != nil {
			return c.JSON(http.StatusBadRequest, "can't query all")
		}
		for row.Next() {
			err := row.Scan(&expense.Id, &expense.Title, &expense.Amount, &expense.Note, pq.Array(&expense.Tags))
			if err != nil {
				return c.JSON(http.StatusBadRequest, err)
			}
			expenses = append(expenses, expense)
		}
		return c.JSON(http.StatusOK, expenses)
	})

	e.POST("/expenses", func(c echo.Context) error {
		err := c.Bind(&expense)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		if errDb != nil {
			return c.JSON(http.StatusBadRequest, errDb)
		}
		row := db.QueryRow("INSERT INTO expenses (title,amount,note,tags) values ($1,$2,$3,$4) RETURNING id ", expense.Title, expense.Amount, expense.Note, pq.Array(expense.Tags))
		var id int
		err = row.Scan(&id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		expense.Id = id
		return c.JSON(http.StatusAccepted, expense)
	})
	e.GET("/expenses/:id", func(c echo.Context) error {
		id := c.Param("id")
		data, errData := db.Prepare("SELECT id,title,amount,note,tags FROM expenses where id=$1")
		if errDb != nil {
			return c.JSON(http.StatusBadRequest, errDb)
		}
		if errData != nil {
			return c.JSON(http.StatusBadRequest, errData)
		}
		row := data.QueryRow(id)
		err := row.Scan(&expense.Id, &expense.Title, &expense.Amount, &expense.Note, pq.Array(&expense.Tags))
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.JSON(http.StatusOK, expense)
	})
	e.PUT("/expenses/:id", func(c echo.Context) error {
		id := c.Param("id")
		err := c.Bind(&expense)
		data, errData := db.Prepare("UPDATE expenses SET title = $2,amount=$3,note=$4,tags=$5 WHERE id=$1;")
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		if errData != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		if _, err := data.Exec(id, expense.Title, expense.Amount, expense.Note, pq.Array(&expense.Tags)); err != nil {
			log.Fatal("error execute update ", err)
			return c.JSON(http.StatusBadRequest, err)
		}
		return c.JSON(http.StatusOK, expense)
	})

	fmt.Println("Please use server.go for main file")
	fmt.Println("start at port:", os.Getenv("PORT"))
	log.Fatal(e.Start(os.Getenv("PORT")))
}

func createTable() error {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS expenses ( id SERIAL PRIMARY KEY, title TEXT, amount FLOAT, note TEXT, tags TEXT[] );")
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
