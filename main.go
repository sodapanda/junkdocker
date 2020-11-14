package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	docker "github.com/fsouza/go-dockerclient"
)

var db *sql.DB
var dockerClient *docker.Client

func main() {
	//连接数据库
	var err error
	db, err = sql.Open("sqlite3", "./junk.db")
	checkErr(err)
	defer db.Close()

	//连接docker
	if dockerClient == nil {
		dockerClient, err = docker.NewClientFromEnv()
		checkErr(err)
	}

	//启动http服务器
	httpServer()
}

func httpServer() {
	http.HandleFunc("/bandwidthReport", ReportBandWidth)
	http.HandleFunc("/start", StartContainer)
	http.ListenAndServe(":1030", nil)
}

//ReportBandWidth 带宽使用记录
func ReportBandWidth(w http.ResponseWriter, r *http.Request) {
	cid := getQueryParam(r, "cid")
	txBytesStr := getQueryParam(r, "txbytes")
	timeStamp := time.Now().Unix()
	txBytes, err := strconv.Atoi(txBytesStr)
	checkErr(err)

	writeDBBandwidth(cid, timeStamp, txBytes)
}

//StartContainer start
func StartContainer(w http.ResponseWriter, r *http.Request) {
	//user id
	uid := getQueryParam(r, "uid")
	//目的ip:port
	dst := getQueryParam(r, "dst")

	//container名称
	ctnrName := fmt.Sprintf("%s_%s", uid, strings.Replace(dst, ":", "_", -1))

	container, err := dockerClient.CreateContainer(docker.CreateContainerOptions{
		Name: ctnrName,
		Config: &docker.Config{
			Image:      "sodapanda/portforward:0.0.1",
			Entrypoint: []string{"./entrypoint.sh", dst, uid},
		},
		HostConfig: &docker.HostConfig{
			PublishAllPorts: true,
			CapAdd:          []string{"NET_ADMIN"},
		},
	})
	checkErr(err)
	log.Printf("created %s %s\n", container.Name, container.ID)

	err = dockerClient.StartContainer(container.ID, nil)
	log.Printf("start %s\n", container.ID)
	checkErr(err)
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

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
