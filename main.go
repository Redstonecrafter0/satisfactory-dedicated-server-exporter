package main

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/collectors"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type SatisfactoryRPCRequest struct {
    Function string `json:"function"`
    Data map[string]string `json:"data"`
}

type SatisfactoryRPCResponseDataServerGameState struct {
    ActiveSessionName string `json:"activeSessionName"`
    NumConnectedPlayers int `json:"numConnectedPlayers"`
    PlayerLimit int `json:"playerLimit"`
    TechTier int `json:"techTier"`
    IsGameRunning bool `json:"isGameRunning"`
    TotalGameDuration int `json:"totalGameDuration"`
    IsGamePaused bool `json:"isGamePaused"`
    AverageTickRate float64 `json:"averageTickRate"`
    AutoLoadSessionName string `json:"autoLoadSessionName"`
}

type SatisfactoryRPCResponseData struct {
    ServerGameState SatisfactoryRPCResponseDataServerGameState `json:"serverGameState"`
}

type SatisfactoryRPDResponse struct {
    Data SatisfactoryRPCResponseData `json:"data"`
}

func main() {
    reg := prometheus.NewRegistry()

    reg.MustRegister(
        NewSatisfactoryCollector(),
        collectors.NewGoCollector(),
        collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
    )

    http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
    log.Fatal(http.ListenAndServe(":8080", nil))
}

type SatisfactoryCollector struct {
    sessionName *prometheus.Desc
    numConnectedPlayers *prometheus.Desc
    playerLimit *prometheus.Desc
    techTier *prometheus.Desc
    isGameRunning *prometheus.Desc
    totalGameDuration *prometheus.Desc
    isGamePaused *prometheus.Desc
    averageTickRate *prometheus.Desc
}

func NewSatisfactoryCollector() *SatisfactoryCollector {
    return &SatisfactoryCollector{
        sessionName: prometheus.NewDesc("satisfactory_dedicated_server_session_name",
            "Name of the currently loaded game session",
            []string{"active", "autoload"}, nil,
        ),
        numConnectedPlayers: prometheus.NewDesc("satisfactory_dedicated_server_num_connected_players",
            "Number of the players currently connected to the Dedicated Server",
            nil, nil,
        ),
        playerLimit: prometheus.NewDesc("satisfactory_dedicated_server_player_limit",
            "Maximum number of players that can be connected to the Dedicated Server",
            nil, nil,
        ),
        techTier: prometheus.NewDesc("satisfactory_dedicated_server_tech_tier",
            "Maximum Tech Tier of all Schematics currently unlocked",
            nil, nil,
        ),
        isGameRunning: prometheus.NewDesc("satisfactory_dedicated_server_is_game_running",
            "1 if the save is currently loaded, 0 if the server is waiting for the session to be created",
            nil, nil,
        ),
        totalGameDuration: prometheus.NewDesc("satisfactory_dedicated_server_total_game_duration",
            "Total time the current save has been loaded, in seconds",
            nil, nil,
        ),
        isGamePaused: prometheus.NewDesc("satisfactory_dedicated_server_is_game_paused",
            "1 if the game is paused. If the game is paused, total game duration does not increase",
            nil, nil,
        ),
        averageTickRate: prometheus.NewDesc("satisfactory_dedicated_server_average_tick_rate",
            "Average tick rate of the server, in ticks per second",
            nil, nil,
        ),
    }
}

func (collector *SatisfactoryCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- collector.sessionName
    ch <- collector.numConnectedPlayers
    ch <- collector.playerLimit
    ch <- collector.techTier
    ch <- collector.isGameRunning
    ch <- collector.totalGameDuration
    ch <- collector.isGamePaused
    ch <- collector.averageTickRate
}

func (collector *SatisfactoryCollector) Collect(ch chan<- prometheus.Metric) {
    data, err := FetchSatisfactoryDedicatedServer()
    if err != nil {
        fmt.Printf("Warning %s\n", err)
        return
    }

    var isGameRunning float64
    if data.IsGameRunning {
        isGameRunning = 1
    } else {
        isGameRunning = 0
    }

    var isGamePaused float64
    if data.IsGamePaused {
        isGamePaused = 1
    } else {
        isGamePaused = 0
    }

    ch <- prometheus.MustNewConstMetric(collector.sessionName, prometheus.GaugeValue, 1, data.ActiveSessionName, data.AutoLoadSessionName)
    ch <- prometheus.MustNewConstMetric(collector.numConnectedPlayers, prometheus.GaugeValue, float64(data.NumConnectedPlayers))
    ch <- prometheus.MustNewConstMetric(collector.playerLimit, prometheus.GaugeValue, float64(data.PlayerLimit))
    ch <- prometheus.MustNewConstMetric(collector.techTier, prometheus.CounterValue, float64(data.TechTier))
    ch <- prometheus.MustNewConstMetric(collector.isGameRunning, prometheus.GaugeValue, isGameRunning)
    ch <- prometheus.MustNewConstMetric(collector.totalGameDuration, prometheus.CounterValue, float64(data.TotalGameDuration))
    ch <- prometheus.MustNewConstMetric(collector.isGamePaused, prometheus.GaugeValue, isGamePaused)
    ch <- prometheus.MustNewConstMetric(collector.averageTickRate, prometheus.GaugeValue, data.AverageTickRate)
}

func FetchSatisfactoryDedicatedServer() (*SatisfactoryRPCResponseDataServerGameState, error) {
    client := http.Client{
        Transport: http.DefaultTransport,
    }
    if os.Getenv("SDSE_INSECURE") == "1" {
        client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
    }
    payload, err := json.Marshal(SatisfactoryRPCRequest{
        Function: "QueryServerState",
        Data: map[string]string{},
    })
    if err != nil {
        return nil, err
    }
    req, err := http.NewRequest("POST", "https://" + os.Getenv("SDSE_HOST") + ":" + os.Getenv("SDSE_PORT") + "/api/v1", bytes.NewReader(payload))
    if err != nil {
        return nil, err
    }
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Authorization", "Bearer " + os.Getenv("SDSE_TOKEN"))
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    buf := bytes.Buffer{}
    _, err = buf.ReadFrom(resp.Body)
    if err != nil {
        return nil, err
    }
    var responseObject SatisfactoryRPDResponse
    err = json.Unmarshal(buf.Bytes(), &responseObject)
    if err != nil {
        return nil, err
    }
    return &responseObject.Data.ServerGameState, nil
}
