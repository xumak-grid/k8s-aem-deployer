package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xumak-grid/go-grid/pkg/client"
	discovery "github.com/xumak-grid/go-grid/pkg/service-discovery"
)

// Pod represents a pod inside the services
type Pod struct {
	Name       string    `json:"name"`
	App        string    `json:"app"`
	Deployment string    `json:"deployment"`
	Runmode    string    `json:"runmode"`
	Status     string    `json:"status"`
	Port       string    `json:"port"`
	Time       time.Time `json:"timestamp"`
}

// deployer represent the status of the deploy to AEM instances
// Every service contains a slice of pods
var deployer map[string][]Pod

//Example request: POST http://deployer-service-url/deploy/{namespace}/{app}/{environment}
func deployHandler(w http.ResponseWriter, r *http.Request) {

	segs := strings.Split(r.URL.Path, "/")
	if len(segs) != 5 {
		jsonError(w, "Bad request, Invalid URL Path", http.StatusBadRequest)
		return
	}
	namespace := segs[2]
	app := segs[3]
	deployment := segs[4]
	selector := map[string]string{"app": app, "deployment": deployment}

	//getting the file from the request
	file, _, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error getting the file", err.Error())
		return
	}

	defer file.Close()
	f, err := os.Create("myPackage.zip")
	if err != nil {
		fmt.Println("Open the new file", err.Error())
		return
	}
	io.Copy(f, file)

	// Get the services that match with the parameters
	clt := client.MustNewKubeClient()
	services, err := discovery.Locate(clt, selector, namespace, false)
	if err != nil {
		fmt.Println("Error getting services", err.Error())
		return
	}

	if len(services) == 0 {
		msg := fmt.Sprintf("No services available in namespace: %s app: %s deployment: %s", namespace, app, deployment)
		fmt.Println(msg)
		jsonError(w, msg, http.StatusBadRequest)
		return
	}

	deployer := make(map[string][]*Pod)
	wg := sync.WaitGroup{}
	wg.Add(len(services))
	mx := sync.Mutex{}

	// Iterate over services to get the their pods
	for _, service := range services {
		go manageService(&wg, service, deployer, &mx)
	}
	wg.Wait()
	encode(w, deployer)
	return

}

func manageService(wg *sync.WaitGroup, service discovery.Service, deployer map[string][]*Pod, mx *sync.Mutex) {
	defer wg.Done()
	podList := []*Pod{}
	ch := make(chan *Pod)
	for _, pd := range service.PodList {
		go managePod(pd, ch)
	}
	for i := 0; i < len(service.PodList); i++ {
		pod := <-ch
		podList = append(podList, pod)
	}
	mx.Lock()
	deployer[service.Name] = podList
	mx.Unlock()
}

func managePod(pd discovery.Pod, ch chan *Pod) {
	ip := pd.PodIP
	newPOD := Pod{}
	newPOD.Name = pd.Name
	newPOD.Time = time.Now()

	//reading labels' pod
	for key, value := range pd.Labels {
		if key == "app" {
			newPOD.App = value
		}
		if key == "deployment" {
			newPOD.Deployment = value
		}
		if key == "runmode" {
			newPOD.Runmode = value
			if value == "author" {
				newPOD.Port = ":4502"
			}
			if value == "publish" {
				newPOD.Port = ":4503"
			}
		}
	}

	//install the package in the pod
	url := "http://" + ip + newPOD.Port + "/crx/packmgr/service.jsp"
	fmt.Printf("Installing package url: %s pod: %s\n", url, newPOD.Name)
	err := installPackage(url, "admin", "admin", "myPackage.zip")
	if err != nil {
		fmt.Println("Error installing the package:", err.Error())
		newPOD.Status = "fail"
	} else {
		newPOD.Status = "success"
	}
	ch <- &newPOD
}
