package main

import (
	"fmt"
	"net/http"

	docker "github.com/fsouza/go-dockerclient"
)

func main() {
	httpServer()
}

func httpServer() {
	http.HandleFunc("/bandwidthReport", ReportBandWidth)
	http.ListenAndServe(":1030", nil)
}

//ReportBandWidth 带宽使用记录
func ReportBandWidth(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
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
