package collector

import (
        "bytes"
        "fmt"
        "sync"
        "time"
        //"math/rand"
        "strconv"
        "github.com/apex/log"
        "github.com/prometheus/client_golang/prometheus"
        gofish "github.com/stmcginnis/gofish"
        gofishcommon "github.com/stmcginnis/gofish/common"
        redfish "github.com/stmcginnis/gofish/redfish"
)

// Metric name parts.
const (
        // Exporter namespace.
        namespace = "redfish"
        // Subsystem(s).
        exporter = "exporter"
        // Math constant for picoseconds to seconds.
        picoSeconds = 1e12
)

type redfishMetric struct {
        desc *prometheus.Desc
}

type Firmware struct {
        Name string
        Version string  //This can be blank for some fields
        Description string
}

// Metric descriptors.
var (
        totalScrapeDurationDesc = prometheus.NewDesc(
                prometheus.BuildFQName(namespace, exporter, "collector_duration_seconds"),
                "Collector time duration.",
                nil, nil,
        )
        firmwareInventoryDesc = prometheus.NewDesc (
                prometheus.BuildFQName("redfish", "system", "firmwareinventory"),
                "Redfish Firmware Inventory Information",
                []string{"hostname", "Id", "name", "version", "description"},
                nil,
        )
)

// RedfishCollector collects redfish metrics. It implements prometheus.Collector.
type RedfishCollector struct {
        redfishClient *gofish.APIClient
        metrics                 map[string]redfishMetric
        collectors    map[string]prometheus.Collector
        redfishUp     prometheus.Gauge
        firmwareInfo []Firmware  
}

// NewRedfishCollector return RedfishCollector
func NewRedfishCollector(host string, username string, password string, logger *log.Entry) *RedfishCollector {
        var collectors map[string]prometheus.Collector
        collectorLogCtx := logger
        redfishClient, err := newRedfishClient(host, username, password)
        if err != nil {
                collectorLogCtx.WithError(err).Error("error creating redfish client")
        } else {
                chassisCollector := NewChassisCollector(namespace, redfishClient, collectorLogCtx)
                systemCollector := NewSystemCollector(namespace, redfishClient, collectorLogCtx)
                managerCollector := NewManagerCollector(namespace, redfishClient, collectorLogCtx)

                collectors = map[string]prometheus.Collector{"chassis": chassisCollector, "system": systemCollector, "manager": managerCollector}
        }
       return &RedfishCollector{
               redfishClient: redfishClient,
               collectors:    collectors,
               redfishUp: prometheus.NewGauge(
                       prometheus.GaugeOpts{
                               Namespace: namespace,
                                Subsystem: "",
                                Name:      "up",
                                Help:      "redfish up",
                        },
                ),
        }
}

// Describe implements prometheus.Collector.
func (r *RedfishCollector) Describe(ch chan<- *prometheus.Desc) {
        for _, collector := range r.collectors {
                collector.Describe(ch)
        }
        ch <- firmwareInventoryDesc

}

// Collect implements prometheus.Collector.
func (r *RedfishCollector) Collect(ch chan<- prometheus.Metric) {

 SystemHostName := ""
 SystemUUID := ""

        scrapeTime := time.Now()
        if r.redfishClient != nil {
                defer r.redfishClient.Logout()
                r.redfishUp.Set(1)
                wg := &sync.WaitGroup{}
                wg.Add(len(r.collectors))
                
                service := r.redfishClient.Service
        // get a list of systems from service
        if systems, err := service.Systems(); err != nil {
                fmt.Println("error getting systems from service", err)
        } else  {
        //PrettyPrint(systems)
        for _, system := range systems {
                SystemHostName = system.HostName
                SystemUUID = system.UUID
}}



                for _, collector := range r.collectors {
                        go func(collector prometheus.Collector) {
                                defer wg.Done()
                                collector.Collect(ch)
                        }(collector)
                }
                
                if SystemHostName == "" {
                        SystemHostName = SystemUUID //This is to avoid a client side failure when parsing Nvme drive firmware information
                }
                        parseFirmwareInventory(ch, r.redfishClient, "/redfish/v1/UpdateService/FirmwareInventory/", SystemHostName) 

                wg.Wait()
        } else {
                r.redfishUp.Set(0)
        }

        ch <- r.redfishUp
        ch <- prometheus.MustNewConstMetric(totalScrapeDurationDesc, prometheus.GaugeValue, time.Since(scrapeTime).Seconds())
}

func parseFirmwareInventory(ch chan<- prometheus.Metric, c gofishcommon.Client,  systemFirmwareInventoryUrl string, systemHostName string) () {
        Id := 0
        fw, err := redfish.ListReferencedSoftwareInventories(c, systemFirmwareInventoryUrl)
         if err != nil {
                fmt.Println(err)
        }
        firmwareSlice := make([]Firmware,0)
        for _, f := range fw {
                        firmwareStruct := Firmware {
                        Name:  f.Name,
                        Version: f.Version,
                        Description: f.Description,
                }
                firmwareSlice = append(firmwareSlice, firmwareStruct)
        }
        for _, firmware := range firmwareSlice {
        //min := 1
        //max := 100
        Id = Id + 1
        //firmware.Name = firmware.Name + "." + strconv.Itoa(rand.Intn(max - min + 1) + min)
        systemFirmwareLabelValues := []string{systemHostName, strconv.Itoa(Id), firmware.Name, firmware.Version, firmware.Description}
        ch <- prometheus.MustNewConstMetric(
                        firmwareInventoryDesc,
                        prometheus.GaugeValue,
                        1,
                        systemFirmwareLabelValues...
                )
    }
}

func newRedfishClient(host string, username string, password string) (*gofish.APIClient, error) {

        url := fmt.Sprintf("https://%s", host)

        config := gofish.ClientConfig{
                Endpoint: url,
                Username: username,
                Password: password,
                Insecure: true,
        }
        redfishClient, err := gofish.Connect(config)
        if err != nil {
                return nil, err
        }
        return redfishClient, nil
}

func parseCommonStatusHealth(status gofishcommon.Health) (float64, bool) {
        if bytes.Equal([]byte(status), []byte("OK")) {
                return float64(1), true
        } else if bytes.Equal([]byte(status), []byte("Warning")) {
                return float64(2), true
        } else if bytes.Equal([]byte(status), []byte("Critical")) {
                return float64(3), true
        }
        return float64(0), false
}


func parseCommonStatusState(status gofishcommon.State) (float64, bool) {

        if bytes.Equal([]byte(status), []byte("")) {
                return float64(0), false
        } else if bytes.Equal([]byte(status), []byte("Enabled")) {
                return float64(1), true
        } else if bytes.Equal([]byte(status), []byte("Disabled")) {
                return float64(2), true
        } else if bytes.Equal([]byte(status), []byte("StandbyOffinline")) {
                return float64(3), true
        } else if bytes.Equal([]byte(status), []byte("StandbySpare")) {
                return float64(4), true
        } else if bytes.Equal([]byte(status), []byte("InTest")) {
                return float64(5), true
        } else if bytes.Equal([]byte(status), []byte("Starting")) {
                return float64(6), true
        } else if bytes.Equal([]byte(status), []byte("Absent")) {
                return float64(7), true
        } else if bytes.Equal([]byte(status), []byte("UnavailableOffline")) {
                return float64(8), true
        } else if bytes.Equal([]byte(status), []byte("Deferring")) {
                return float64(9), true
        } else if bytes.Equal([]byte(status), []byte("Quiesced")) {
                return float64(10), true
        } else if bytes.Equal([]byte(status), []byte("Updating")) {
                return float64(11), true
        }
        return float64(0), false
}


func parseCommonPowerState(status redfish.PowerState) (float64, bool) {
        if bytes.Equal([]byte(status), []byte("On")) {
                return float64(1), true
        } else if bytes.Equal([]byte(status), []byte("Off")) {
                return float64(2), true
        } else if bytes.Equal([]byte(status), []byte("PoweringOn")) {
                return float64(3), true
        } else if bytes.Equal([]byte(status), []byte("PoweringOff")) {
                return float64(4), true
        }
        return float64(0), false
}

func parseLinkStatus(status redfish.LinkStatus) (float64, bool) {
        if bytes.Equal([]byte(status), []byte("LinkUp")) {
                return float64(1), true
        } else if bytes.Equal([]byte(status), []byte("NoLink")) {
                return float64(2), true
        } else if bytes.Equal([]byte(status), []byte("LinkDown")) {
                return float64(3), true
        }
        return float64(0), false
}

func parsePortLinkStatus(status redfish.PortLinkStatus) (float64, bool) {
        if bytes.Equal([]byte(status), []byte("Up")) {
                return float64(1), true
        } 
        return float64(0), false
}
func boolToFloat64(data bool) float64 {

        if data {
                return float64(1)
        }
        return float64(0)

}

func parsePhySecReArmMethod(method redfish.IntrusionSensorReArm) (float64, bool) {
        if bytes.Equal([]byte(method), []byte("Manual")) {
                return float64(1), true
        }
        if bytes.Equal([]byte(method), []byte("Automatic")) {
                return float64(2), true
        }

        return float64(0), false
}

func parsePhySecIntrusionSensor(method redfish.IntrusionSensor) (float64, bool) {
        if bytes.Equal([]byte(method), []byte("Normal")) {
                return float64(1), true
        }
        if bytes.Equal([]byte(method), []byte("TamperingDetected")) {
                return float64(2), true
        }
        if bytes.Equal([]byte(method), []byte("HardwareIntrusion")) {
                return float64(3), true
        }

        return float64(0), false
}
