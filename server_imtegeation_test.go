//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Response struct {
	*http.Response
	err error
}

func TestCreateData(t *testing.T) {
	var expense Expenses
	body := bytes.NewBufferString(`{
	"title": "strawberry smoothie",
    "amount": 90,
    "note": "night market promotion discount 10 bath", 
    "tags": ["food", "beverage"]
	}`)
	res := request(http.MethodPost, uri("expenses"), body)
	err := res.Decode(&expense)
	if err != nil {
		t.Fatal("can't create expense")
	}
	assert.Nil(t, err)
	assert.EqualValues(t, http.StatusAccepted, res.StatusCode)
	assert.NotEqual(t, 0, expense.Id)
	assert.Equal(t, "strawberry smoothie", expense.Title)
	assert.Equal(t, 90, expense.Amount)
	assert.Equal(t, "night market promotion discount 10 bath", expense.Note)
	assert.Equal(t, []string{"food", "beverage"}, expense.Tags)
}
func TestGetExpenseAll(t *testing.T) {
	var expenses []Expenses
	res := request(http.MethodGet, uri("expenses"), nil)
	err := res.Decode(&expenses)

	assert.Nil(t, err)
	assert.EqualValues(t, http.StatusOK, res.StatusCode)
	assert.Greater(t, len(expenses), 0)
}
func TestGetExpenseById(t *testing.T) {
	var expense Expenses
	res := request(http.MethodGet, uri("expenses", strconv.Itoa(1)), nil)
	err := res.Decode(&expense)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, 1, expense.Id)
	assert.NotEmpty(t, expense.Title)
	assert.NotEmpty(t, expense.Amount)
	assert.NotEmpty(t, expense.Note)
	assert.NotEmpty(t, expense.Tags)
}
func TestUpdateExpenseById(t *testing.T) {
	body := bytes.NewBufferString(`{
		"title": "apple smoothie",
		"amount": 89,
		"note": "no discount",
		"tags": ["beverage"]
		}`)
	var expense Expenses
	res := request(http.MethodPut, uri("expenses", strconv.Itoa(1)), body)
	err := res.Decode(&expense)
	if err != nil {
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	}
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, 1, expense.Id)
	assert.NotEmpty(t, expense.Title)
	assert.NotEmpty(t, expense.Amount)
	assert.NotEmpty(t, expense.Note)
	assert.NotEmpty(t, expense.Tags)
}

func (r *Response) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}
	return json.NewDecoder(r.Body).Decode(v)
}
func uri(paths ...string) string {
	host := "http://localhost:2565"
	if paths == nil {
		return host
	}
	url := append([]string{host}, paths...)
	return strings.Join(url, "/")
}

func request(method, url string, body io.Reader) *Response {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("Authorization", "Basic U2F5ZmFyOjEyMzQ=")
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	return &Response{res, err}
}
