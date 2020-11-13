package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"

	docker "github.com/fsouza/go-dockerclient"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./junk.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	httpServer()
}

func httpServer() {
	http.HandleFunc("/bandwidthReport", ReportBandWidth)
	http.ListenAndServe(":1030", nil)
}

//ReportBandWidth 带宽使用记录
func ReportBandWidth(w http.ResponseWriter, r *http.Request) {
	cid := getQueryParam(r, "cid")
	txBytesStr := getQueryParam(r, "txbytes")
	timeStamp := time.Now().Unix()
	txBytes, err := strconv.Atoi(txBytesStr)
	if err != nil {
		log.Fatal(err)
	}

	writeDBBandwidth(cid, timeStamp, txBytes)
}

func testDocker() {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}
	imgs, err := client.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		panic(err)
	}
	for _, img := range imgs {
		fmt.Println("ID: ", img.ID)
		fmt.Println("RepoTags: ", img.RepoTags)
		fmt.Println("Created: ", img.Created)
		fmt.Println("Size: ", img.Size)
		fmt.Println("VirtualSize: ", img.VirtualSize)
		fmt.Println("ParentId: ", img.ParentID)
		fmt.Println("====================================")
	}
}

func getQueryParam(r *http.Request, queryKey string) string {
	return r.URL.Query()[queryKey][0]
}

func writeDBBandwidth(containerID string, timeStamp int64, txBytes int) {
	sqlStmt := fmt.Sprintf("insert into bandwidth(ContainerID,Timestamp,TxBytes) values('%s',%d,%d)",
		containerID, timeStamp, txBytes)
	fmt.Println(sqlStmt)
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatal(err)
	}
}
