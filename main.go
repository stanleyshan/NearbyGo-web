package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"strconv"
)

type Location struct { //type like class in java
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Post struct {
	// `json:"user"` is for the json parsing of this User field. Otherwise, by default it's 'User'.
	User     string `json:"user"`
	Message  string  `json:"message"`
	Location Location `json:"location"`
}

const (
	DISTANCE = "200km"
)

func main() {
	fmt.Println("started-service")
	http.HandleFunc("/post", handlerPost) //如果發來的帖子是在post這個url底下，就用handlerPost這個函數去處理
	http.HandleFunc("/search", handlerSearch)
	log.Fatal(http.ListenAndServe(":8080", nil)) //start 這個service，並且在8080端口監聽他  //Fatal是如果返回是error的話，就返回一個Fatal級別的log
}

func handlerPost(w http.ResponseWriter, r *http.Request) { //1.22.00   //r *http.Request 收到從29行傳來的資料
	// Parse from body of request to get a json object.
	fmt.Println("Received one post request")
	decoder := json.NewDecoder(r.Body) //r.Body: postman 裡面 post的body
	var p Post
	if err := decoder.Decode(&p); err != nil { //當錯誤不為空，代表有錯
		panic(err)
		return
	}

	fmt.Fprintf(w, "Post received: %s\n", p.Message)
}


func handlerSearch(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one post request")
	lat := r.URL.Query().Get("lat") //to get request parameters from url  //lat是以string格式寫出來的float number
	lon := r.URL.Query().Get("lon")

	lt, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64) //解析: 將lat(String)轉成64位的float ->???????????
	ln, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	ran := DISTANCE
	//lat=10&lon=100&range=300，就會從200km -> 300km
	if val := r.URL.Query().Get("range"); val != "" {
		ran = val + "km"
	}

	fmt.Printf("Search received: %s %s %s", lat, lon, ran) //在console上顯示

	// Return a fake post
	p := &Post{ //&Post，意思是用指針抓住這個post的object，p代表是地址的感覺
		User:"1111",
		Message:"一生必去的100个地方",
		Location: Location{
			Lat:lt,
			Lon:ln,
		},
	}

	js, err := json.Marshal(p) //Marshal，將Go的object轉成JSON的String格式 -> JSON以byte表示???????????????
	if err != nil {
		panic(err)
		return
	}

	w.Header().Set("Content-Type", "application/json") //返回的時候告訴瀏覽器這是json object
	w.Write(js) //因為已經是byte array所以可以直接寫回response裡去
}