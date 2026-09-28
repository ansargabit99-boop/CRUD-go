package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)
type ErrorResponse struct {
	Message string `json:"message"`
}
type todo struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Done bool `json:"done"`
}
var (
	mu sync.Mutex
	todos = []todo{{ID:1,Title:"learn go"}}
	nextId = 2
)

func getTodo(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-type","application/json")
	json.NewEncoder(w).Encode(todos)
}
func createTodo(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if err := 	json.NewDecoder(r.Body).Decode(&input) ;err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid json"})
		return
	}
	if input.Title == "" {
		w.Header().Set("Content-type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "the title must not be empty"})
		return
	}
	mu.Lock()
	newtodo:= todo{ID:nextId,Title: input.Title}
	nextId++
	todos = append(todos, newtodo)
	mu.Unlock()
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newtodo)
}
func changeDone(w http.ResponseWriter, r *http.Request) {
	var id  = r.PathValue("id")


	idNum,err:= strconv.Atoi(id)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message:"invalid id"})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	for i,t:= range todos {
		if t.ID == idNum {
			todos[i].Done = !todos[i].Done
			w.Header().Set("Content-Type","application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(t)
			return
		}
	}
	w.Header().Set("Content-type","application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(ErrorResponse{Message: "not found"})

}
func deleteOne(w http.ResponseWriter,r *http.Request) {
	 var id = r.PathValue("id")
	 
	 idNum,err:= strconv.Atoi(id)
	 if err != nil {
		w.Header().Set("Content-Type","applciation/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message:"invalid id"})
	 }

	 for i,t:= range todos {
		if t.ID==idNum {
			todos = append(todos[:i], todos[i+1:]...)
			w.Header().Set("Content-Type","application/json")
			w.WriteHeader(http.StatusNoContent)
			return
		}
	 }
	 w.Header().Set("Content-Type","application/json")
	 w.WriteHeader(http.StatusNotFound)
	 json.NewEncoder(w).Encode(ErrorResponse{Message:"not found"})
}
func getOne(w http.ResponseWriter,r *http.Request) {
	var id = r.PathValue("id")
	idNum,err:=strconv.Atoi(id)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message:"invaild id"})
		return
	}
	mu.Lock()
    defer mu.Unlock()

	for _,t:=range todos {
		if t.ID == idNum {
			w.Header().Set("Content-Type","application/json")
			json.NewEncoder(w).Encode(t)
			return
		}
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(ErrorResponse{Message: "not found"})
}
func main() {
	http.HandleFunc("GET /todos",getTodo)
	http.HandleFunc("POST /todos",createTodo)
	http.HandleFunc("GET /todos/{id}",getOne)
	http.HandleFunc("PATCH /todos{id}",changeDone)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
