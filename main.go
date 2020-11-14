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
	//各个容器每10秒上报一次流量(这10秒内的流量)
	http.HandleFunc("/bandwidthReport", ReportBandWidth)
	//启动一个容器的接口
	http.HandleFunc("/start", StartContainer)
	//获取数据量
	http.HandleFunc("/getTxBytes", GetBytesInTime)
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
	fmt.Fprintf(w, "{\"ok\":%s}", "true")
}

//StartContainer start
func StartContainer(w http.ResponseWriter, r *http.Request) {
	//user id
	uid := getQueryParam(r, "uid")
	//目的ip:port
	dst := getQueryParam(r, "dst")

	//container名称
	ctnrName := fmt.Sprintf("%s_%s", uid, strings.Replace(dst, ":", "_", -1))

	//创建container
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

	//启动container
	err = dockerClient.StartContainer(container.ID, nil)
	log.Printf("start %s\n", container.ID)
	checkErr(err)

	//端口映射
	ctnOut, err := dockerClient.InspectContainerWithOptions(docker.InspectContainerOptions{
		ID: container.ID,
	})
	mappedPort := ctnOut.NetworkSettings.Ports["8800/tcp"][0].HostPort
	fmt.Fprintf(w, "{\"ok\":true,\"id\":\"%s\",\"port\":\"%s\"}", container.ID, mappedPort)
}

//GetBytesInTime 获取一段时间内的流量
func GetBytesInTime(w http.ResponseWriter, r *http.Request) {
	cid := getQueryParam(r, "cid")
	startTs, err := strconv.Atoi(getQueryParam(r, "start"))
	checkErr(err)
	endTs, err := strconv.Atoi(getQueryParam(r, "end"))
	checkErr(err)

	sqlStmt := fmt.Sprintf("select * from bandwidth where ContainerID='%s' and Timestamp > %d and Timestamp < %d", cid, startTs, endTs)
	rows, err := db.Query(sqlStmt)

	total := 0
	for rows.Next() {
		var cid string
		var timestamp int
		var txBytes int

		err = rows.Scan(&cid, &timestamp, &txBytes)
		checkErr(err)
		total = total + txBytes
	}
	defer rows.Close()

	fmt.Fprintf(w, "{\"cid\":\"%s\",\"start\":\"%d\",\"end\":\"%d\",\"txBytes\":\"%d\"}", cid, startTs, endTs, total)
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
