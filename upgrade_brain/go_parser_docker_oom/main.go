package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	oomKilledMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "docker_container_oom_killed",
			Help: "1 if the container was killed due to OOM, 0 otherwise",
		},
		[]string{"container_name"},
	)
)

func init() {
	prometheus.MustRegister(oomKilledMetric)
}

func getContainers() ([]string, error) {
	cmd := exec.Command("/usr/bin/sudo", "/usr/bin/docker", "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	containers := strings.Split(strings.TrimSpace(string(output)), "\n")
	return containers, nil
}

func checkOOMKilled(containerName string) {
	cmd := exec.Command("/usr/bin/sudo", "/usr/bin/docker", "inspect", "--format", "{{.State.OOMKilled}}", containerName)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error executing command for container %s: %v", containerName, err)
		return
	}

	oomKilledStr := strings.TrimSpace(string(output))
	var oomKilledValue float64
	if oomKilledStr == "true" {
		oomKilledValue = 1
		cmd := exec.Command("/usr/bin/sudo", "/usr/bin/docker", "restart", containerName)
		if err := cmd.Run(); err != nil {
			log.Printf("Ошибка при перезапуске контейнера %s: %v", containerName, err)
		}
	} else {
		oomKilledValue = 0
	}

	oomKilledMetric.WithLabelValues(containerName).Set(oomKilledValue)
}

func monitorContainers() {
	for {
		containers, err := getContainers()
		if err != nil {
			log.Printf("Error fetching containers: %v", err)
			time.Sleep(60 * time.Second)
			continue
		}

		for _, container := range containers {
			checkOOMKilled(container)
		}

		time.Sleep(60 * time.Second)
	}
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

func main() {
	go monitorContainers()

	http.HandleFunc("/metrics", metricsHandler)
	fmt.Println("Starting server on :9099")
	log.Fatal(http.ListenAndServe(":9099", nil))
}

