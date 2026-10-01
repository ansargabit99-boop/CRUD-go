package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)
type Server struct {
	pool *pgxpool.Pool
}
type ErrorResponse struct{
	Message string
}
type succesData struct{
	token string
	name string
}
func connectDb() (*pgxpool.Pool,error){
	db,err := pgxpool.New(context.Background(),os.Getenv("DATABASE_URL"))
	if err!= nil {
		return nil,err
	}
	err = db.Ping(context.Background())
	if err !=nil {
		return  nil,err
	}

	return db,err
}
func loadingEnv(){
	err:=godotenv.Load()
	if err!= nil {
		log.Fatal("error while loadin enviroment variables")
	}
}
func generateToken(id int) (string,error) {
	claims:= jwt.MapClaims{
		"userId":id,
		"exp": time.Now().Add(30*24*time.Hour).Unix(),
	}
	token:=jwt.NewWithClaims(jwt.SigningMethodES256,claims)
	
	secret:=os.Getenv("JWT_SECRET")
	return  token.SignedString([]byte(secret))
}
func(s *Server) RegisterHandle(w http.ResponseWriter, r *http.Request){
	pool:=s.pool
	type sendedData struct {
		Name string `json:"name"`
		Gmail string `json:"gmail"`
		Number string `json:"number"`
	}
	var userData = sendedData{}
	err:= json.NewDecoder(r.Body).Decode(&userData)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message:"invalid data"})
		return
	} 
	type returnedData struct {
		Id int `json:"id"`
		Name string `json:"name"`
	}
	var returned = returnedData{}
	err = pool.QueryRow(r.Context(),"INSERT INTO users (name,gmail,number) VALUES($1,$2,$3) RETURNING id,name",userData.Name,
	userData.Gmail,
	userData.Number).Scan(&returned.Id,returned.Name)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "something went wrong"})
		return
	}
	token,err:=generateToken(returned.Id)
	if err !=nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "something went wrong"})
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(succesData{token:token,name:returned.Name})

}
func(s*Server) loginHandle(w http.ResponseWriter,r *http.Request) {
	pool:=s.pool
	type details struct {
		NumOrGmail string `json:"numOrGmail"`
		Password string `json:"password"`
	}
	var loginDet = details{}
	err:= json.NewDecoder(r.Body).Decode(&loginDet)
	if err!= nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message:"something went wrong"})
		return
	}
	type data struct {
		Id int `json:"id"`
		Name string `json:"name"`
		Gmail string `json:"gmail"`
		Number string `json:"number"`
	}
	var rowDet = data{}
	if strings.Contains(loginDet.NumOrGmail,"@") {
	    err := pool.QueryRow(r.Context(),"SELECT id,name,gmail,number FROM users WHERE gmail = $1",loginDet.NumOrGmail).Scan(&rowDet.Id,&rowDet.Name,&rowDet.Gmail,&rowDet.Number)
		if err != nil {
			w.Header().Set("Content-Type","application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "something went wrong"})
			return
		}
	} else {
         err := pool.QueryRow(r.Context(),"SELECT id,name,gmail,number FROM users WHERE number = $1",loginDet.NumOrGmail).Scan(&rowDet.Id,&rowDet.Name,&rowDet.Gmail,&rowDet.Number)
		 if err != nil {
			w.Header().Set("Content-Type","application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "something went wrong"})
			return
		 }
	}
	token,err := generateToken(rowDet.Id)
	if err != nil {
		w.Header().Set("Content-Type","application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "unauthoised"})
		return
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(succesData{token: token,name:rowDet.Name})
	
}
func main() {
	loadingEnv()
	pool,err:= connectDb()
	if err != nil {
		log.Fatal("something went wrong while connection")
	}
	server := &Server{
		pool: pool,
	}
	defer pool.Close()
	port:=":8080"
	http.HandleFunc("POST /register",server.RegisterHandle)
	http.HandleFunc("/POST /login",server.loginHandle)
	log.Fatal(http.ListenAndServe(port,http.DefaultServeMux))	

}