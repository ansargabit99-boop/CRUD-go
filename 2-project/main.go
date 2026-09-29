package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)
type ErrorResponse struct {
    Message string `json:"message"`
}
type Server struct {
	pool *pgxpool.Pool
}
type Users struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Gmail string `json:"gmail"`
	Number string `json:"number"`
}
type SuccesMessage struct {
	Message string `json:"message"`
}
func connectDb() (*pgxpool.Pool,error) {
	ctx:=context.Background()
	pool,err:=pgxpool.New(ctx,"postgres://postgres:ansGAB2009.@localhost:5432/users")
	if err!=nil {
		fmt.Println(err)
		return nil, err
	}
	return pool,nil	
}
func (s *Server) getUsers(w http.ResponseWriter,r *http.Request) {
	var pool = s.pool
	var data []Users
	users,err := pool.Query(r.Context(),"SELECT * FROM users")
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(	ErrorResponse{Message:"something went wrong"})
		return
	}
	defer users.Close()
	for users.Next() {
		var user Users
		
		err:= users.Scan(&user.Id,&user.Name,&user.Gmail,&user.Number)
		if err != nil {
			return
		}
		data = append(data, user)

	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(data)
}
func(s*Server) DeleteUser(w http.ResponseWriter,r *http.Request) {
	var userId = r.PathValue("id")
	var pool = s.pool
	userIdNum,err:= strconv.Atoi(userId)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message:"Invalid id"})
		return
	}
	_,err = pool.Query(r.Context(),"DELETE FROM users WHERE id=$1",userIdNum)
	if err !=nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message:"something went wrong"})
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(SuccesMessage{Message: "the user succesfully deleted"})
	
}
func(s*Server) AddUser(w http.ResponseWriter,r *http.Request) {
	pool:=s.pool
	type addedUsers struct {
		Name string `json:"name"`
		Gmail string `json:"gmail"`
		Number string `json:"number"`
	}
	var datas addedUsers
	err:= json.NewDecoder(r.Body).Decode(&datas)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid json"})
		return
	}
	var user Users
	
	err = pool.QueryRow(
		r.Context(),
		"INSERT INTO users (name, gmail, number) VALUES ($1, $2, $3) RETURNING id, name, gmail, number",
		datas.Name,
		datas.Gmail,
		datas.Number,
	).Scan(
		&user.Id,
		&user.Name,
		&user.Gmail,
		&user.Number,
	)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "something went wrong"})
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(user)
	
}

func(s *Server) changeSomething(w http.ResponseWriter, r *http.Request){
	var pool = s.pool
	var userId = r.PathValue("id")
	type body struct {
		Name string `json:"name"`
		Gmail string `json:"gmail"`
		Number string `json:"number"`
	}
	var items body
	err := json.NewDecoder(r.Body).Decode(&items)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message:"Invalid body"})
	}
	Query := "UPDATE users SET "
	var change = []string{}
	var values = []any{}
	i:=1
	userIdNum,err := strconv.Atoi(userId)
	if err!=nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid Id"})
		return
	}
	if items.Name!="" {
		change = append(change, fmt.Sprintf("name=$%d", i))
		values = append(values, items.Name)
        i++
	}
	if items.Gmail != "" {
		change = append(change, fmt.Sprintf("gmail=$%d",i))
		values= append(values, items.Gmail)
		i++
	}
	if items.Number != "" {
		change = append(change, fmt.Sprintf("number=$%d",i))
		values = append(values, items.Number)
		i++
	}
	if len(change) == 0 {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Nothing To Change"})
		return
	}
	final := strings.Join(change, ",")
	Query+=final
	values = append(values, userIdNum)
	Query+= fmt.Sprintf(" WHERE id=$%d RETURNING *",i)
	var newUser Users
	rows,err := pool.Query(r.Context(),Query,values...)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "something went wrong"})
		return
	}
	defer rows.Close()
	rows.Scan(&newUser.Id,&newUser.Name,&newUser.Gmail,&newUser.Number)
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(newUser)

}
func main() {
    pool,err := connectDb()
	if err!= nil {
		log.Fatal("something went wrong")
	}
	defer pool.Close()
	server:=Server{
		pool: pool,
	}
	http.HandleFunc("GET /users", server.getUsers)
	log.Fatal(http.ListenAndServe(":8080",nil))
}