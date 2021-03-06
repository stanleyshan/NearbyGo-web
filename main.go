package main

import (
	elastic "gopkg.in/olivere/elastic.v3"
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"strconv"
	"reflect"
	"github.com/pborman/uuid"
	"strings"
	//"context"
        //"cloud.google.com/go/bigtable"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"context"
	"cloud.google.com/go/storage"

	"io"
)

var mySigningKey = []byte("secret")  ////這是你的private key，不能到github

type Location struct { //type like class in java
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Post struct {
	// `json:"user"` is for the json parsing of this User field. Otherwise, by default it's 'User'.
	User     string `json:"user"`
	Message  string  `json:"message"`
	Location Location `json:"location"`
	Url    string `json:"url"`
}

const (
	INDEX = "around"
	TYPE = "post"
	DISTANCE = "200km"
	// Needs to update
	PROJECT_ID = "around-194517"
	BT_INSTANCE = "around-post"
	// Needs to update this URL if you deploy it to cloud.
	ES_URL = "http://35.231.117.16:9200"
	// Needs to update this bucket based on your gcs bucket name.
	BUCKET_NAME = "post-images-194517"

)

func main() {
	// Create a client
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(INDEX).Do()
	if err != nil {
		panic(err)
	}
	if !exists {
		// Create a new index. //將location的類型變成geo_point，elastic才會知道lat與lon是地理位置
		mapping := `{
                    "mappings":{
                           "post":{
                                  "properties":{
                                         "location":{
                                                "type":"geo_point"
                                         }
                                  }
                           }
                    }
             }
             `
		_, err := client.CreateIndex(INDEX).Body(mapping).Do()
		if err != nil {
			// Handle error
			panic(err)
		}
	}

	fmt.Println("Started service successfully")

	// Here we are instantiating the gorilla/mux router
	r := mux.NewRouter()  //加一個中間層，讓我們可以檢測jwt這個token

	var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return mySigningKey, nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	r.Handle("/post", jwtMiddleware.Handler(http.HandlerFunc(handlerPost))).Methods("POST")
	r.Handle("/search", jwtMiddleware.Handler(http.HandlerFunc(handlerSearch))).Methods("GET")
	r.Handle("/login", http.HandlerFunc(loginHandler)).Methods("POST")
	r.Handle("/signup", http.HandlerFunc(signupHandler)).Methods("POST")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))

	//fmt.Println("started-service")
	//http.HandleFunc("/post", handlerPost) //如果發來的帖子是在post這個url底下，就用handlerPost這個函數去處理
	//http.HandleFunc("/search", handlerSearch)
	//log.Fatal(http.ListenAndServe(":8080", nil)) //start 這個service，並且在8080端口監聽他  //Fatal是如果返回是error的話，就返回一個Fatal級別的log

	////會卡在log?
	//m := make(map[string]bool) //key: string value: bool
	//m["jack"] = true
	//m["john"] = false
	//for k, v := range m {
	//	fmt.Println(k, v)
	//}
	//
	//l := []string{"a", "b", "c"}
	//for _, v := range l { //因為range返回的是index, value，我們這邊不care index所以用_表示
	//	fmt.Println(v)
	//}


}

/*func handlerPost(w http.ResponseWriter, r *http.Request) { //1.22.00   //r *http.Request 收到從29行傳來的資料
	// Parse from body of request to get a json object.
	fmt.Println("Received one post request")
	decoder := json.NewDecoder(r.Body)
	var p Post
	if err := decoder.Decode(&p); err != nil {
		panic(err)
		return
	}

	// Create a client
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	id := uuid.New()
	user := r.Context().Value("user")
	claims := user.(*jwt.Token).Claims
	username := claims.(jwt.MapClaims)["username"]
	p.User = username.(string)

	// Save it to index
	_, err = es_client.Index().
		Index(INDEX).
		Type(TYPE).
		Id(id).
		BodyJson(p).
		Refresh(true).
		Do()
	if err != nil {
		panic(err)
		return
	}

	fmt.Printf("Post is saved to Index: %s\n", p.Message)
	/*
	ctx := context.Background()
	bt_client, err := bigtable.NewClient(ctx, PROJECT_ID, BT_INSTANCE)
	if err != nil {
		panic(err)
		return
	}

	tbl := bt_client.Open("post")
	mut := bigtable.NewMutation()
	t := bigtable.Now()
	mut.Set("post", "user", t, []byte(p.User))
	mut.Set("post", "message", t, []byte(p.Message))
	mut.Set("location", "lat", t, []byte(strconv.FormatFloat(p.Location.Lat, 'f', -1, 64)))
	mut.Set("location", "lon", t, []byte(strconv.FormatFloat(p.Location.Lon, 'f', -1, 64)))

	err = tbl.Apply(ctx, id, mut)
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("Post is saved to BigTable: %s\n", p.Message)
	*/
//} //127行

	/*
	// Parse from body of request to get a json object.
	fmt.Println("Received one post request")
	decoder := json.NewDecoder(r.Body) //r.Body: postman 裡面 post的body
	var p Post
	if err := decoder.Decode(&p); err != nil { //當錯誤不為空，代表有錯
		panic(err)
		return
	}

	//fmt.Fprintf(w, "Post received: %s\n", p.Message)

	id := uuid.New()
	// Save to ES.
	saveToES(&p, id)

	//Save to BigTable
	ctx := context.Background()
	// you must update project name here
	bt_client, err := bigtable.NewClient(ctx, PROJECT_ID, BT_INSTANCE)
	if err != nil {
		panic(err)
		return
	}

	// TODO (student questions) save Post into BT as well

	tbl := bt_client.Open("post")
	mut := bigtable.NewMutation()
	t := bigtable.Now()

	mut.Set("post", "user", t, []byte(p.User))
	mut.Set("post", "message", t, []byte(p.Message))
	mut.Set("location", "lat", t, []byte(strconv.FormatFloat(p.Location.Lat, 'f', -1, 64)))
	mut.Set("location", "lon", t, []byte(strconv.FormatFloat(p.Location.Lon, 'f', -1, 64)))

	err = tbl.Apply(ctx, id, mut)
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("Post is saved to BigTable: %s\n", p.Message)

}*/


// Save a post to ElasticSearch
func saveToES(p *Post, id string) {
	// Create a client
	es_client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Save it to index
	_, err = es_client.Index(). //創建index
		Index(INDEX).
		Type(TYPE).
		Id(id).
		BodyJson(p).
		Refresh(true).
		Do() //提交
	if err != nil {
		panic(err)
		return
	}

	fmt.Printf("Post is saved to Index: %s\n", p.Message)
}



func handlerSearch(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received one post request")
	//lat := r.URL.Query().Get("lat") //to get request parameters from url  //lat是以string格式寫出來的float number
	//lon := r.URL.Query().Get("lon")

	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64) //解析: 將lat(String)轉成64位的float ->???????????
	lon, _ := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	ran := DISTANCE
	//lat=10&lon=100&range=300，就會從200km -> 300km
	if val := r.URL.Query().Get("range"); val != "" {
		ran = val + "km"
	}

	fmt.Printf( "Search received: %f %f %s\n", lat, lon, ran)

	// Create a client
	client, err := elastic.NewClient(elastic.SetURL(ES_URL), elastic.SetSniff(false))
	if err != nil {
		panic(err)
		return
	}

	// Define geo distance query as specified in
	// https://www.elastic.co/guide/en/elasticsearch/reference/5.2/query-dsl-geo-distance-query.html
	q := elastic.NewGeoDistanceQuery("location")
	q = q.Distance(ran).Lat(lat).Lon(lon)

	// Some delay may range from seconds to minutes. So if you don't get enough results. Try it later.
	searchResult, err := client.Search().
		Index(INDEX).
		Query(q).
		Pretty(true).
		Do()
	if err != nil {
		// Handle error
		panic(err)
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	// TotalHits is another convenience function that works even when something goes wrong.
	fmt.Printf("Found a total of %d post\n", searchResult.TotalHits())

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization.
	var typ Post
	var ps []Post
	for _, item := range searchResult.Each(reflect.TypeOf(typ)) { // instance of
		p := item.(Post) // p = (Post) item
		fmt.Printf("Post by %s: %s at lat %v and lon %v\n", p.User, p.Message, p.Location.Lat, p.Location.Lon)
		// TODO(student homework): Perform filtering based on keywords such as web spam etc.
		if !containsFilteredWords(&p.Message) {
			ps = append(ps, p)
		}
	}
	js, err := json.Marshal(ps)
	if err != nil {
		panic(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(js)


	//fmt.Printf("Search received: %s %s %s", lat, lon, ran) //在console上顯示
	//
	//// Return a fake post
	//p := &Post{ //&Post，意思是用指針抓住這個post的object，p代表是地址的感覺
	//	User:"1111",
	//	Message:"一生必去的100个地方",
	//	Location: Location{
	//		Lat:lt,
	//		Lon:ln,
	//	},
	//}
	//
	//js, err := json.Marshal(p) //Marshal，將Go的object轉成JSON的String格式 -> JSON以byte表示???????????????
	//if err != nil {
	//	panic(err)
	//	return
	//}
	//
	//w.Header().Set("Content-Type", "application/json") //返回的時候告訴瀏覽器這是json object
	//w.Write(js) //因為已經是byte array所以可以直接寫回response裡去
}

func containsFilteredWords(s *string) bool {
	filteredWords := []string{
		"fuck",
		"200",
	}
	for _, word := range filteredWords {
		if strings.Contains(*s, word) {
			return true
		}
	}
	return false
}

//Lesson54 Google Cloud Storage
func handlerPost(w http.ResponseWriter, r *http.Request) {
	// Other codes
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")


	// 32 << 20 is the maxMemory param for ParseMultipartForm, equals to 32MB (1MB = 1024 * 1024 bytes = 2^20 bytes)
	// After you call ParseMultipartForm, the file will be saved in the server memory with maxMemory size.
	// If the file size is larger than maxMemory, the rest of the data will be saved in a system temporary file.
	r.ParseMultipartForm(32 << 20)

	// Parse from form data.
	fmt.Printf("Received one post request %s\n", r.FormValue("message"))
	lat, _ := strconv.ParseFloat(r.FormValue("lat"), 64)
	lon, _ := strconv.ParseFloat(r.FormValue("lon"), 64)
	p := &Post{
		User:    "1111",
		Message: r.FormValue("message"),
		Location: Location{
			Lat: lat,
			Lon: lon,
		},
	}

	id := uuid.New()

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image is not available", http.StatusInternalServerError)
		fmt.Printf("Image is not available %v.\n", err)
		return
	}
	defer file.Close()

	ctx := context.Background()

	// replace it with your real bucket name.
	_, attrs, err := saveToGCS(ctx, file, BUCKET_NAME, id)
	if err != nil {
		http.Error(w, "GCS is not setup", http.StatusInternalServerError)
		fmt.Printf("GCS is not setup %v\n", err)
		return
	}

	// Update the media link after saving to GCS.
	p.Url = attrs.MediaLink

	// Save to ES.
	saveToES(p, id)

	// Save to BigTable.
	//saveToBigTable(p, id)

}

func saveToGCS(ctx context.Context, r io.Reader, bucket, name string) (*storage.ObjectHandle, *storage.ObjectAttrs, error) {
	// Student questions

	//Creates a client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close() //函數完成後，不管出錯與否，都要把資源關閉

	//Creates a Bucket instance.
	bh := client.Bucket(bucket)
	// Next check if the bucket exists
	if _, err = bh.Attrs(ctx); err != nil {
		return nil, nil, err //oblecthandle, objectattribute, error //go語言出錯返回的寫法
	}

	//把內容寫進Bucket
	obj := bh.Object(name)    //獲得object的內容
	w := obj.NewWriter(ctx)   //創建新的writter
	if _, err := io.Copy(w, r); err != nil {   //拷貝文件
		return nil, nil, err
	}
	if err := w.Close(); err != nil {
		return nil, nil, err
	}

	//設定存取權限
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil { //讓所有用戶都有可讀的權限
		return nil, nil, err
	}

	attrs, err := obj.Attrs(ctx)
	fmt.Printf("Post is saved to GCS: %s\n", attrs.MediaLink) //輸出link告訴我們文件上傳到哪裡去了
	return obj, attrs, err

}
