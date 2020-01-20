package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// Pod represents a pod inside the services
type Pod struct {
	Name    string    `json:"name"`
	Runmode string    `json:"runmode"`
	Port    string    `json:"port"`
	Status  string    `json:"status"`
	Time    time.Time `json:"timestamp"`
}

func main() {
	fmt.Println("sendHandler")
	url := ""
	fileToSend := ""
	var b bytes.Buffer
	wri := multipart.NewWriter(&b)

	file, err := os.Open(fileToSend)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	form, err := wri.CreateFormFile("file", fileToSend)
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, err = io.Copy(form, file); err != nil {
		fmt.Println(err)
		return
	}
	wri.Close()
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", wri.FormDataContentType())
	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	//fmt.Println(string(body))

	var services map[string][]Pod

	err = json.Unmarshal(body, &services)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(string(body))
		os.Exit(1)
	}

	for name, pods := range services {
		fmt.Println("Service Name:", name)
		for _, pd := range pods {
			fmt.Printf("Instance %s info:\n", pd.Name)
			fmt.Println("  Runmode:", pd.Runmode)
			fmt.Println("  Port:", pd.Port)
			fmt.Println("  Status:", pd.Status)
			fmt.Println("  Timestamp", pd.Time)
		}
	}

	return
}
