package collector

import (
	"fmt"
	"sync"
	//"math/rand"
	"strconv"
    //"time"
	"encoding/json"
	"github.com/apex/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/stmcginnis/gofish/oem/hpe"
)

// A SystemCollector implements the prometheus.Collector.

type systemMetric struct {
	desc *prometheus.Desc
}

func PrettyPrint(data interface{}) {
    var p []byte
    //    var err := error
    p, err := json.MarshalIndent(data, "", "\t")
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Printf("%s \n", p)
}

// SystemSubsystem is the system subsystem
var (
	SystemSubsystem                   = "system"
	SystemLabelNames                  = []string{"hostname", "resource", "system_id"}
	SystemModelLabelNames             = []string{"hostname", "resource", "asset_tag", "bios_version", "description", "manufacturer", "model", "sku", "serial_number", "uuid"}
	SystemMemoryLabelNames            = []string{"hostname", "resource", "memory", "memory_id"}
	SystemProcessorLabelNames         = []string{"hostname", "resource", "processor", "processor_id"}
	SystemVolumeLabelNames            = []string{"hostname", "resource", "volume", "volume_id"}
	SystemDeviceLabelNames            = []string{"hostname", "resource", "device"}
	SystemDriveLabelNames             = []string{"hostname", "resource", "drive", "drive_id"}
	SystemStorageControllerLabelNames = []string{"hostname", "resource", "storage_controller", "storage_controller_id"}
	SystemPCIeDeviceLabelNames        = []string{"hostname", "resource", "Id", "pcie_device", "pcie_device_id", "pcie_device_partnumber", "pcie_device_type", "pcie_serial_number"}
	SystemNetworkInterfaceLabelNames  = []string{"hostname", "resource", "network_interface", "network_interface_id"}
	SystemEthernetInterfaceLabelNames = []string{"hostname", "resource", "ethernet_interface", "ethernet_interface_id", "ethernet_interface_speed"}
	SystemUsageLabelNames             = []string{"hostname", "resource"}
	//"avgcpu0freq", "avgcpu1freq", "cpu0power", "cpu1power", "cpuicutil", "cpuutil", "iobusutil", "jittercount", "memorybustil"}
	SystemPCIeFunctionLabelNames      = []string{"hostname", "resource", "pcie_function_name", "pcie_function_id", "pci_function_deviceclass", "pci_function_type"}

	systemMetrics = map[string]systemMetric{
		"system_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "state"),
				"system state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemLabelNames,
				nil,
			),
		},
		"system_model_info": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "model_info"),
        "organization responsible for producing the system, the name by which the manufacturer generally refers to the system, and a part number and sku assigned by the organization that is responsible for producing or manufacturing the system",
        SystemModelLabelNames,
				nil,
			),
		},
		"system_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "health_state"),
				"system health,1(OK),2(Warning),3(Critical)",
				SystemLabelNames,
				nil,
			),
		},
		"system_power_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "power_state"),
				"system power state",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_memory_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "total_memory_state"),
				"system overall memory state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_memory_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "total_memory_health_state"),
				"system overall memory health,1(OK),2(Warning),3(Critical)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_memory_size": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "total_memory_size"),
				"system total memory size, GiB",
				SystemLabelNames,
				nil,
			),
		},
		/////Jay adding new fields for HP OEM specific ones
		//"AvgCPU0Freq", "AvgCPU1Freq", "CPU0Power", "CPU1Power", "CPUICUtil", "CPUUtil", "IOBusUtil", "JitterCount", "MemoryBusUtil"
		"system_systemusage": {
			desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "system_usage"),
        "Oem specific metrics from a HPE system providing system usage details",
         SystemUsageLabelNames,
                                nil,
                        ),
                },
		"system_systemusage_avg_cpu0_freq": {
                        desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "systemusage_avgcpu0freq"),
                                "average CPU0 frequency, MHz",
                                SystemUsageLabelNames,
                                nil,
                        ),
                },
                "system_systemusage_avg_cpu1_freq": {
                        desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "systemusage_avgcpu1freq"),
                                "average CPU1 frequency, MHz",
                                SystemUsageLabelNames,
                                nil,
                        ),
                },
                "system_systemusage_cpu0_power": {
                        desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "systemusage_cpu0power"),
                                "system CPU0 power in watts",
                                SystemUsageLabelNames,
                                nil,
                        ),
                },
                "system_systemusage_cpu1_power": {
                        desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "systemusage_cpu1power"),
                                "system CPU1 power in watts",
                                SystemUsageLabelNames,
                                nil,
                        ),
                },
                "system_systemusage_cpu_icutil": {
                        desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "systemusage_cpuicutil"),
                                "system CPU IC util",
                                SystemUsageLabelNames,
                                nil,
                        ),
                },
                "system_systemusage_cpu_util": {
                        desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "systemusage_cpuutil"),
                                "system CPU util",
                                SystemUsageLabelNames,
                                nil,
                        ),
                },
                "system_systemusage_io_bus_util": {
                        desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "systemusage_iobusutil"),
                                "system io bus util",
                                SystemUsageLabelNames,
                                nil,
                        ),
                },
                "system_systemusage_jitter_count": {
                        desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "systemusage_jittercount"),
                                "system jitter count",
                                SystemUsageLabelNames,
                                nil,
                        ),
                },
                "system_systemusage_memory_bus_util": {
                        desc: prometheus.NewDesc(
                                prometheus.BuildFQName(namespace, SystemSubsystem, "systemusage_memorybusutil"),
                                "system memory bus util",
                                SystemUsageLabelNames,
                                nil,
                        ),
                },
		"system_total_processor_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "total_processor_state"),
				"system overall processor state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_processor_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "total_processor_health_state"),
				"system overall processor health,1(OK),2(Warning),3(Critical)",
				SystemLabelNames,
				nil,
			),
		},
		"system_total_processor_count": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "total_processor_count"),
				"system total  processor count",
				SystemLabelNames,
				nil,
			),
		},
		"system_memory_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "memory_state"),
				"system memory state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemMemoryLabelNames,
				nil,
			),
		},
		"system_memory_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "memory_health_state"),
				"system memory  health state,1(OK),2(Warning),3(Critical)",
				SystemMemoryLabelNames,
				nil,
			),
		},
		"system_memory_capacity": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "memory_capacity"),
				"system memory capacity, MiB",
				SystemMemoryLabelNames,
				nil,
			),
		},

		"system_processor_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_state"),
				"system processor state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_processor_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_health_state"),
				"system processor  health state,1(OK),2(Warning),3(Critical)",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_processor_total_threads": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_total_threads"),
				"system processor total threads",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_processor_total_cores": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "processor_total_cores"),
				"system processor total cores",
				SystemProcessorLabelNames,
				nil,
			),
		},
		"system_simple_storage_device_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "simple_storage_device_state"),
				"system simple storage device state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemDeviceLabelNames,
				nil,
			),
		},
		"system_simple_storage_device_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "simple_storage_device_health_state"),
				"system simple storage device health state,1(OK),2(Warning),3(Critical)",
				SystemDeviceLabelNames,
				nil,
			),
		},
		"system_storage_volume_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_volume_state"),
				"system storage volume state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemVolumeLabelNames,
				nil,
			),
		},
		"system_storage_volume_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_volume_health_state"),
				"system storage volume health state,1(OK),2(Warning),3(Critical)",
				SystemVolumeLabelNames,
				nil,
			),
		},
		"system_storage_volume_capacity": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_volume_capacity"),
				"system storage volume capacity,Bytes",
				SystemVolumeLabelNames,
				nil,
			),
		},
		"system_storage_drive_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_drive_state"),
				"system storage drive state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemDriveLabelNames,
				nil,
			),
		},
		"system_storage_drive_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_drive_health_state"),
				"system storage volume health state,1(OK),2(Warning),3(Critical)",
				SystemDriveLabelNames,
				nil,
			),
		},
		"system_storage_drive_capacity": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_drive_capacity"),
				"system storage drive capacity,Bytes",
				SystemDriveLabelNames,
				nil,
			),
		},
		"system_storage_controller_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_controller_state"),
				"system storage controller state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemStorageControllerLabelNames,
				nil,
			),
		},
		"system_storage_controller_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "storage_controller_health_state"),
				"system storage controller health state,1(OK),2(Warning),3(Critical)",
				SystemStorageControllerLabelNames,
				nil,
			),
		},
		"system_pcie_device_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "pcie_device_state"),
				"system pcie device state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemPCIeDeviceLabelNames,
				nil,
			),
		},
		"system_pcie_device_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "pcie_device_health_state"),
				"system pcie device health state,1(OK),2(Warning),3(Critical)",
				SystemPCIeDeviceLabelNames,
				nil,
			),
		},
		"system_pcie_function_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "pcie_function_state"),
				"system pcie device state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemPCIeFunctionLabelNames,
				nil,
			),
		},
		"system_pcie_function_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "pcie_function_health_state"),
				"system pcie device health state,1(OK),2(Warning),3(Critical)",
				SystemPCIeFunctionLabelNames,
				nil,
			),
		},
		"system_network_interface_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "network_interface_state"),
				"system network interface state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemNetworkInterfaceLabelNames,
				nil,
			),
		},
		"system_network_interface_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "network_interface_health_state"),
				"system network interface health state,1(OK),2(Warning),3(Critical)",
				SystemNetworkInterfaceLabelNames,
				nil,
			),
		},
		"system_ethernet_interface_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "ethernet_interface_state"),
				"system ethernet interface state,1(Enabled),2(Disabled),3(StandbyOffinline),4(StandbySpare),5(InTest),6(Starting),7(Absent),8(UnavailableOffline),9(Deferring),10(Quiesced),11(Updating)",
				SystemEthernetInterfaceLabelNames,
				nil,
			),
		},
		"system_ethernet_interface_health_state": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "ethernet_interface_health_state"),
				"system ethernet interface health state,1(OK),2(Warning),3(Critical)",
				SystemEthernetInterfaceLabelNames,
				nil,
			),
		},
		"system_ethernet_interface_link_status": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "ethernet_interface_link_status"),
				"system ethernet interface link status，1(LinkUp),2(NoLink),3(LinkDown)",
				SystemEthernetInterfaceLabelNames,
				nil,
			),
		},
		"system_ethernet_interface_link_enabled": {
			desc: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, SystemSubsystem, "ethernet_interface_link_enabled"),
				"system ethernet interface if the link is enabled",
				SystemEthernetInterfaceLabelNames,
				nil,
			),
		},
	}
)

// SystemCollector implemented prometheus.Collector
type SystemCollector struct {
	redfishClient           *gofish.APIClient
	metrics                 map[string]systemMetric
	collectorScrapeStatus   *prometheus.GaugeVec
	collectorScrapeDuration *prometheus.SummaryVec
	Log                     *log.Entry
}

// NewSystemCollector returns a collector that collecting memory statistics
func NewSystemCollector(namespace string, redfishClient *gofish.APIClient, logger *log.Entry) *SystemCollector {
	return &SystemCollector{
		redfishClient: redfishClient,
		metrics:       systemMetrics,
		Log: logger.WithFields(log.Fields{
			"collector": "SystemCollector",
		}),
		collectorScrapeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "collector_scrape_status",
				Help:      "collector_scrape_status",
			},
			[]string{"collector"},
		),
	}
}

// Describe implements prometheus.Collector.
func (s *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range s.metrics {
		ch <- metric.desc
	}
	s.collectorScrapeStatus.Describe(ch)

}

// Collect implements prometheus.Collector.
func (s *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	collectorLogContext := s.Log
	//get service
	service := s.redfishClient.Service
	PcieDeviceID := 0

	// get a list of systems from service
	if systems, err := service.Systems(); err != nil {
		collectorLogContext.WithField("operation", "service.Systems()").WithError(err).Error("error getting systems from service")
	} else {
		for _, system := range systems {
			systemLogContext := collectorLogContext.WithField("System", system.ID)
			systemLogContext.Info("collector scrape started")
			// overall system metrics

			SystemID := system.ID
			systemHostName := system.HostName
			systemPowerState := system.PowerState
			systemState := system.Status.State
			systemHealthState := system.Status.Health
			systemTotalProcessorCount := system.ProcessorSummary.Count
			systemTotalProcessorsState := system.ProcessorSummary.Status.State
			systemTotalProcessorsHealthState := system.ProcessorSummary.Status.Health
			systemTotalMemoryState := system.MemorySummary.Status.State
			systemTotalMemoryHealthState := system.MemorySummary.Status.Health
			systemTotalMemoryAmount := system.MemorySummary.TotalSystemMemoryGiB
//"AvgCPU0Freq", "AvgCPU1Freq", "CPU0Power", "CPU1Power", "CPUICUtil", "CPUUtil", "IOBusUtil", "JitterCount", "MemoryBusUtil"
           //Need if Manufacturer is HPE, then collect the OEM specific fields listed below else continue
		   //ComuterSystemOem is from the gofish library under gofish/oem/hpe/computersystem.go
		   if (system.Manufacturer == "HPE") { 
        //This is defined in hpe module in the lib 
			wg10 := &sync.WaitGroup{}
                        wg10.Add(1)
                        go parseOemFunction(ch, systemHostName, system, wg10)
		   }
			systemAssetTag := system.AssetTag
			systemBIOSVersion := system.BIOSVersion
			systemDescription := system.Description
      systemManufacturer := system.Manufacturer
      systemModel := system.Model
      systemSKU := system.SKU
      systemSerialNumber := system.SerialNumber
      systemUUID := system.UUID
      systemModelLabelValues := []string{systemHostName, "system", systemAssetTag, systemBIOSVersion, systemDescription, systemManufacturer, systemModel, systemSKU, systemSerialNumber, systemUUID}
			ch <- prometheus.MustNewConstMetric(s.metrics["system_model_info"].desc, prometheus.GaugeValue, 1, systemModelLabelValues...)

			systemLabelValues := []string{systemHostName, "system", SystemID}
			if systemHealthStateValue, ok := parseCommonStatusHealth(systemHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_health_state"].desc, prometheus.GaugeValue, systemHealthStateValue, systemLabelValues...)
			}
			if systemStateValue, ok := parseCommonStatusState(systemState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_state"].desc, prometheus.GaugeValue, systemStateValue, systemLabelValues...)
			}
			if systemPowerStateValue, ok := parseCommonPowerState(systemPowerState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_power_state"].desc, prometheus.GaugeValue, systemPowerStateValue, systemLabelValues...)

			}
			if systemTotalProcessorsStateValue, ok := parseCommonStatusState(systemTotalProcessorsState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_state"].desc, prometheus.GaugeValue, systemTotalProcessorsStateValue, systemLabelValues...)

			}
			if systemTotalProcessorsHealthStateValue, ok := parseCommonStatusHealth(systemTotalProcessorsHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_health_state"].desc, prometheus.GaugeValue, systemTotalProcessorsHealthStateValue, systemLabelValues...)

			}
			ch <- prometheus.MustNewConstMetric(s.metrics["system_total_processor_count"].desc, prometheus.GaugeValue, float64(systemTotalProcessorCount), systemLabelValues...)

			if systemTotalMemoryStateValue, ok := parseCommonStatusState(systemTotalMemoryState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_state"].desc, prometheus.GaugeValue, systemTotalMemoryStateValue, systemLabelValues...)

			}
			if systemTotalMemoryHealthStateValue, ok := parseCommonStatusHealth(systemTotalMemoryHealthState); ok {
				ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_health_state"].desc, prometheus.GaugeValue, systemTotalMemoryHealthStateValue, systemLabelValues...)

			}
			ch <- prometheus.MustNewConstMetric(s.metrics["system_total_memory_size"].desc, prometheus.GaugeValue, float64(systemTotalMemoryAmount), systemLabelValues...)

			// get system OdataID
			//systemOdataID := system.ODataID

			// process memory metrics
			// construct memory Link
			//memoriesLink := fmt.Sprintf("%sMemory/", systemOdataID)

			//if memories, err := redfish.ListReferencedMemorys(s.redfishClient, memoriesLink); err != nil {
			memories, err := system.Memory()
			if err != nil {
				systemLogContext.WithField("operation", "system.Memory()").WithError(err).Error("error getting memory data from system")
			} else if memories == nil {
				systemLogContext.WithField("operation", "system.Memory()").Info("no memory data found")
			} else {
				wg1 := &sync.WaitGroup{}
				wg1.Add(len(memories))

				for _, memory := range memories {
					go parseMemory(ch, systemHostName, memory, wg1)

				}
			}

			// process processor metrics

			//processorsLink := fmt.Sprintf("%sProcessors/", systemOdataID)

			//if processors, err := redfish.ListReferencedProcessors(s.redfishClient, processorsLink); err != nil {
			processors, err := system.Processors()
			if err != nil {
				systemLogContext.WithField("operation", "system.Processors()").WithError(err).Error("error getting processor data from system")
			} else if processors == nil {
				systemLogContext.WithField("operation", "system.Processors()").Info("no processor data found")
			} else {
				wg2 := &sync.WaitGroup{}
				wg2.Add(len(processors))

				for _, processor := range processors {
					go parsePorcessor(ch, systemHostName, processor, wg2)

				}

			}

			//process storage
			//storagesLink := fmt.Sprintf("%sStorage/", systemOdataID)

			//if storages, err := redfish.ListReferencedStorages(s.redfishClient, storagesLink); err != nil {
			storages, err := system.Storage()
			if err != nil {
				systemLogContext.WithField("operation", "system.Storage()").WithError(err).Error("error getting storage data from system")
			} else if storages == nil {
				systemLogContext.WithField("operation", "system.Storage()").Info("no storage data found")
			} else {
				for _, storage := range storages {
					if volumes, err := storage.Volumes(); err != nil {
						systemLogContext.WithField("operation", "system.Volumes()").WithError(err).Error("error getting storage data from system")
					} else {
						wg3 := &sync.WaitGroup{}
						wg3.Add(len(volumes))

						for _, volume := range volumes {
							go parseVolume(ch, systemHostName, volume, wg3)
						}
					}

					drives, err := storage.Drives()
					if err != nil {
						systemLogContext.WithField("operation", "system.Drives()").WithError(err).Error("error getting drive data from system")
					} else if drives == nil {
						systemLogContext.WithFields(log.Fields{"operation": "system.Drives()", "storage": storage.ID}).Info("no drive data found")
					} else {
						wg4 := &sync.WaitGroup{}
						wg4.Add(len(drives))
						for _, drive := range drives {
							go parseDrive(ch, systemHostName, drive, wg4)
						}
					}

					//					if storagecontrollers, err := storage.StorageControllers(); err != nil {
					//						log.Infof("Errors Getting storagecontrollers from system storage : %s", err)
					//					} else {
					//
					//						for _, controller := range storagecontrollers {
					//
					//							controllerODataIDslice := strings.Split(controller.ODataID, "/")
					//							controllerName := controllerODataIDslice[len(controllerODataIDslice)-1]
					//							controllerState := controller.Status.State
					//							controllerHealthState := controller.Status.Health
					//							controllerLabelValues := []string{ "storage_controller", controllerName, systemHostName)
					//							if controllerStateValue,ok := parseCommonStatusState(controllerState); ok {
					//								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_state"].desc, prometheus.GaugeValue, controllerStateValue, //controllerLabelValues...)
					//
					//							}
					//							if controllerHealthStateValue,ok := parseCommonStatusHealth(controllerHealthState); ok {
					//								ch <- prometheus.MustNewConstMetric(s.metrics["system_storage_controller_health_state"].desc, prometheus.GaugeValue, controllerHealthStateValue, //controllerLabelValues...)
					//
					//							}
					//
					//						}
					//
					//					}

				}
			}
			//process pci devices
			//pciDevicesLink := fmt.Sprintf("%sPcidevice/", systemOdataID)
			pcieDevices, err := system.PCIeDevices()
			if err != nil {
				systemLogContext.WithField("operation", "system.PCIeDevices()").WithError(err).Error("error getting PCI-E device data from system")
			} else if pcieDevices == nil {
				systemLogContext.WithField("operation", "system.PCIeDevices()").Info("no PCI-E device data found")
			} else {
				wg5 := &sync.WaitGroup{}
				wg5.Add(len(pcieDevices))
				for _, pcieDevice := range pcieDevices {
					PcieDeviceID = PcieDeviceID + 1
					go parsePcieDevice(ch, systemHostName, pcieDevice, wg5, PcieDeviceID)
				}
			}

			//process networkinterfaces
			networkInterfaces, err := system.NetworkInterfaces()
			if err != nil {
				systemLogContext.WithField("operation", "system.NetworkInterfaces()").WithError(err).Error("error getting network interface data from system")
			} else if networkInterfaces == nil {
				systemLogContext.WithField("operation", "system.NetworkInterfaces()").Info("no network interface data found")
			} else {
				wg6 := &sync.WaitGroup{}
				wg6.Add(len(networkInterfaces))
				for _, networkInterface := range networkInterfaces {
					go parseNetworkInterface(ch, systemHostName, networkInterface, wg6)
				}

			}

			//process ethernetinterfaces
			ethernetInterfaces, err := system.EthernetInterfaces()
			if err != nil {
				systemLogContext.WithField("operation", "system.EthernetInterfaces()").WithError(err).Error("error getting ethernet interface data from system")
			} else if ethernetInterfaces == nil {
				systemLogContext.WithField("operation", "system.PCIeDevices()").Info("no ethernet interface data found")
			} else {
				wg7 := &sync.WaitGroup{}
				wg7.Add(len(ethernetInterfaces))
				for _, ethernetInterface := range ethernetInterfaces {
					go parseEthernetInterface(ch, systemHostName, ethernetInterface, wg7)
				}
			}

			//process simple storage
			simpleStorages, err := system.SimpleStorages()
			if err != nil {
				systemLogContext.WithField("operation", "system.SimpleStorages()").WithError(err).Error("error getting simple storage data from system")
			} else if simpleStorages == nil {
				systemLogContext.WithField("operation", "system.SimpleStorages()").Info("no simple storage data found")
			} else {
				for _, simpleStorage := range simpleStorages {
					devices := simpleStorage.Devices
					wg8 := &sync.WaitGroup{}
					wg8.Add(len(devices))
					for _, device := range devices {
						go parseDevice(ch, systemHostName, device, wg8)
					}
				}
			}
			//process pci functions
			pcieFunctions, err := system.PCIeFunctions()
			if err != nil {
				systemLogContext.WithField("operation", "system.PCIeFunctions()").WithError(err).Error("error getting PCI-E device function data from system")
			} else if pcieFunctions == nil {
				systemLogContext.WithField("operation", "system.PCIeFunctions()").Info("no PCI-E device function data found")
			} else {
				wg9 := &sync.WaitGroup{}
				wg9.Add(len(pcieFunctions))
				for _, pcieFunction := range pcieFunctions {
					go parsePcieFunction(ch, systemHostName, pcieFunction, wg9)
				}
			}
			systemLogContext.Info("collector scrape completed")
		}
		s.collectorScrapeStatus.WithLabelValues("system").Set(float64(1))
	}
}

func parseOemFunction(ch chan<- prometheus.Metric, systemHostName string, system *redfish.ComputerSystem, wg *sync.WaitGroup) {
//Oem.Hpe.SystemUsage":{"AvgCPU0Freq":28,"AvgCPU1Freq":23,"CPU0Power":61,"CPU1Power":52,"CPUICUtil":0,"CPUUtil":0,"IOBusUtil":0,"JitterCount":1597,"MemoryBusUtil":1}
        defer wg.Done()
        //oemFunctionName := fmt.Sprint("%v", 
        //PrettyPrint(system)
	oem, err := hpe.FromComputerSystem(system)
        //This is defined in hpe module in the lib
         if err != nil {
                panic(err)
        }
        //PrettyPrint(oem.Oem.Hpe.SystemUsage)
        avgcpu0freq :=  oem.Oem.Hpe.SystemUsage.AvgCPU0Freq
        avgcpu1freq :=  oem.Oem.Hpe.SystemUsage.AvgCPU1Freq
        cpu0power :=  oem.Oem.Hpe.SystemUsage.CPU0Power
        cpu1power :=   oem.Oem.Hpe.SystemUsage.CPU1Power
        cpuicutil :=   oem.Oem.Hpe.SystemUsage.CPUICUtil
        cpuutil :=   oem.Oem.Hpe.SystemUsage.CPUUtil
        iobusutil :=  oem.Oem.Hpe.SystemUsage.IOBusUtil
        jittercount :=  oem.Oem.Hpe.SystemUsage.JitterCount
        memorybusutil :=  oem.Oem.Hpe.SystemUsage.MemoryBusUtil
       // fmt.Printf("%d, %d, %d, %d, %d, %d, %d, %d, %d", avgcpu0freq, avgcpu1freq,  cpu0power, cpu1power, cpuicutil, cpuutil,  iobusutil, jittercount, memorybusutil)
      //systemOemFunctionLabelLabelValues := []int{avgcpu0freq, avgcpu1freq,  cpu0power, cpu1power, cpuicutil, cpuutil,  iobusutil, jittercount, memorybusutil}
       //PrettyPrint(systemOemFunctionLabelLabelValues)
       systemOemLabelValues := []string{systemHostName, "system_usage"}
	//[]string{systemHostName,  avgcpu0freq, avgcpu1freq,  cpu0power, cpu1power, cpuicutil, cpuutil,  iobusutil, jittercount, memorybusutil}
        //PrettyPrint(systemOemLabelValues)
       // if pciFunctionStateValue, ok := parseCommonStatusState(pciFunctionState); ok {
       ch <- prometheus.MustNewConstMetric(systemMetrics["system_systemusage_avg_cpu0_freq"].desc, prometheus.GaugeValue, float64(avgcpu0freq), systemOemLabelValues...)
       ch <- prometheus.MustNewConstMetric(systemMetrics["system_systemusage_avg_cpu1_freq"].desc, prometheus.GaugeValue, float64(avgcpu1freq), systemOemLabelValues...)
       ch <- prometheus.MustNewConstMetric(systemMetrics["system_systemusage_cpu0_power"].desc, prometheus.GaugeValue, float64(cpu0power), systemOemLabelValues...)
       ch <- prometheus.MustNewConstMetric(systemMetrics["system_systemusage_cpu1_power"].desc, prometheus.GaugeValue, float64(cpu1power), systemOemLabelValues...)
       ch <- prometheus.MustNewConstMetric(systemMetrics["system_systemusage_cpu_icutil"].desc, prometheus.GaugeValue, float64(cpuicutil), systemOemLabelValues...)
       ch <- prometheus.MustNewConstMetric(systemMetrics["system_systemusage_cpu_util"].desc, prometheus.GaugeValue, float64(cpuutil), systemOemLabelValues...)
       ch <- prometheus.MustNewConstMetric(systemMetrics["system_systemusage_io_bus_util"].desc, prometheus.GaugeValue, float64(iobusutil), systemOemLabelValues...)
       ch <- prometheus.MustNewConstMetric(systemMetrics["system_systemusage_jitter_count"].desc, prometheus.GaugeValue, float64(jittercount), systemOemLabelValues...)
       ch <- prometheus.MustNewConstMetric(systemMetrics["system_systemusage_memory_bus_util"].desc, prometheus.GaugeValue, float64(memorybusutil), systemOemLabelValues...)
        //}
        //if pciFunctionHealthStateValue, ok := parseCommonStatusHealth(pciFunctionHealthState); ok {
        //        ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_function_health_state"].desc, prometheus.GaugeValue, pciFunctionHealthStateValue, systemPCIeFunctionLabelLabelValues...)
        //}
}
func parseMemory(ch chan<- prometheus.Metric, systemHostName string, memory *redfish.Memory, wg *sync.WaitGroup) {
	defer wg.Done()
	memoryName := memory.Name
	memoryID := memory.ID
	//memoryDeviceLocator := memory.DeviceLocator
	memoryCapacityMiB := memory.CapacityMiB
	memoryState := memory.Status.State
	memoryHealthState := memory.Status.Health

	systemMemoryLabelValues := []string{systemHostName, "memory", memoryName, memoryID}
	if memoryStateValue, ok := parseCommonStatusState(memoryState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_state"].desc, prometheus.GaugeValue, memoryStateValue, systemMemoryLabelValues...)

	}
	if memoryHealthStateValue, ok := parseCommonStatusHealth(memoryHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_health_state"].desc, prometheus.GaugeValue, memoryHealthStateValue, systemMemoryLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_memory_capacity"].desc, prometheus.GaugeValue, float64(memoryCapacityMiB), systemMemoryLabelValues...)

}

func parsePorcessor(ch chan<- prometheus.Metric, systemHostName string, processor *redfish.Processor, wg *sync.WaitGroup) {
	defer wg.Done()
	processorName := processor.Name
	processorID := processor.ID
	processorTotalCores := processor.TotalCores
	processorTotalThreads := processor.TotalThreads
	processorState := processor.Status.State
	processorHelathState := processor.Status.Health

	systemProcessorLabelValues := []string{systemHostName, "processor", processorName, processorID}

	if processorStateValue, ok := parseCommonStatusState(processorState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_state"].desc, prometheus.GaugeValue, processorStateValue, systemProcessorLabelValues...)

	}
	if processorHelathStateValue, ok := parseCommonStatusHealth(processorHelathState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_health_state"].desc, prometheus.GaugeValue, processorHelathStateValue, systemProcessorLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_threads"].desc, prometheus.GaugeValue, float64(processorTotalThreads), systemProcessorLabelValues...)
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_processor_total_cores"].desc, prometheus.GaugeValue, float64(processorTotalCores), systemProcessorLabelValues...)
}
func parseVolume(ch chan<- prometheus.Metric, systemHostName string, volume *redfish.Volume, wg *sync.WaitGroup) {
	defer wg.Done()
	volumeName := volume.Name
	volumeID := volume.ID
	volumeCapacityBytes := volume.CapacityBytes
	volumeState := volume.Status.State
	volumeHealthState := volume.Status.Health
	systemVolumeLabelValues := []string{systemHostName, "volume", volumeName, volumeID}
	if volumeStateValue, ok := parseCommonStatusState(volumeState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volume_state"].desc, prometheus.GaugeValue, volumeStateValue, systemVolumeLabelValues...)

	}
	if volumeHealthStateValue, ok := parseCommonStatusHealth(volumeHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volume_health_state"].desc, prometheus.GaugeValue, volumeHealthStateValue, systemVolumeLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_volume_capacity"].desc, prometheus.GaugeValue, float64(volumeCapacityBytes), systemVolumeLabelValues...)
}
func parseDevice(ch chan<- prometheus.Metric, systemHostName string, device redfish.Device, wg *sync.WaitGroup) {
	defer wg.Done()
	deviceName := device.Name
	deviceState := device.Status.State
	deviceHealthState := device.Status.Health
	systemDeviceLabelValues := []string{systemHostName, "device", deviceName}
	if deviceStateValue, ok := parseCommonStatusState(deviceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_simple_storage_device_state"].desc, prometheus.GaugeValue, deviceStateValue, systemDeviceLabelValues...)
	}
	if deviceHealthStateValue, ok := parseCommonStatusHealth(deviceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_simple_storage_device_health_state"].desc, prometheus.GaugeValue, deviceHealthStateValue, systemDeviceLabelValues...)
	}
}
func parseDrive(ch chan<- prometheus.Metric, systemHostName string, drive *redfish.Drive, wg *sync.WaitGroup) {
	defer wg.Done()
	driveName := drive.Name
	driveID := drive.ID
	driveCapacityBytes := drive.CapacityBytes
	driveState := drive.Status.State
	driveHealthState := drive.Status.Health
	systemdriveLabelValues := []string{systemHostName, "drive", driveName, driveID}
	if driveStateValue, ok := parseCommonStatusState(driveState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_state"].desc, prometheus.GaugeValue, driveStateValue, systemdriveLabelValues...)

	}
	if driveHealthStateValue, ok := parseCommonStatusHealth(driveHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_health_state"].desc, prometheus.GaugeValue, driveHealthStateValue, systemdriveLabelValues...)

	}
	ch <- prometheus.MustNewConstMetric(systemMetrics["system_storage_drive_capacity"].desc, prometheus.GaugeValue, float64(driveCapacityBytes), systemdriveLabelValues...)
}

func parsePcieDevice(ch chan<- prometheus.Metric, systemHostName string, pcieDevice *redfish.PCIeDevice, wg *sync.WaitGroup, Id int) {

	defer wg.Done()
	pcieDeviceName := pcieDevice.Name
	pcieDeviceID := pcieDevice.ID
//pcieDevice.ID can have same value from some Dell servers. To fix this, adding a random number at the end to make them uniq
    //rand.Seed(time.Now().UnixNano())
    //min := 1
    //max := 100
	//pcieDeviceID := pcieDevice.ID + strconv.Itoa(rand.Intn(max - min + 1) + min)

	pcieDeviceState := pcieDevice.Status.State
	pcieDeviceHealthState := pcieDevice.Status.Health
	pcieDevicePartNumber := pcieDevice.PartNumber
	pcieDeviceType := fmt.Sprintf("%v,", pcieDevice.DeviceType)
	pcieSerialNumber := pcieDevice.SerialNumber
	systemPCIeDeviceLabelValues := []string{systemHostName, strconv.Itoa(Id), "pcie_device", pcieDeviceName, pcieDeviceID, pcieDevicePartNumber, pcieDeviceType, pcieSerialNumber}

	if pcieStateVaule, ok := parseCommonStatusState(pcieDeviceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_device_state"].desc, prometheus.GaugeValue, pcieStateVaule, systemPCIeDeviceLabelValues...)

	}
	if pcieHealthStateVaule, ok := parseCommonStatusHealth(pcieDeviceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_device_health_state"].desc, prometheus.GaugeValue, pcieHealthStateVaule, systemPCIeDeviceLabelValues...)

	}
}

func parseNetworkInterface(ch chan<- prometheus.Metric, systemHostName string, networkInterface *redfish.NetworkInterface, wg *sync.WaitGroup) {
	defer wg.Done()
	networkInterfaceName := networkInterface.Name
	networkInterfaceID := networkInterface.ID
	networkInterfaceState := networkInterface.Status.State
	networkInterfaceHealthState := networkInterface.Status.Health
	systemNetworkInterfaceLabelValues := []string{systemHostName, "network_interface", networkInterfaceName, networkInterfaceID}

	if networknetworkInterfaceStateVaule, ok := parseCommonStatusState(networkInterfaceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_network_interface_state"].desc, prometheus.GaugeValue, networknetworkInterfaceStateVaule, systemNetworkInterfaceLabelValues...)

	}
	if networknetworkInterfaceHealthStateVaule, ok := parseCommonStatusHealth(networkInterfaceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_network_interface_health_state"].desc, prometheus.GaugeValue, networknetworkInterfaceHealthStateVaule, systemNetworkInterfaceLabelValues...)

	}
}

func parseEthernetInterface(ch chan<- prometheus.Metric, systemHostName string, ethernetInterface *redfish.EthernetInterface, wg *sync.WaitGroup) {
	defer wg.Done()
	//ethernetInterfaceODataIDslice := strings.Split(ethernetInterface.ODataID, "/")
	//ethernetInterfaceName := ethernetInterfaceODataIDslice[len(ethernetInterfaceODataIDslice)-1]
	ethernetInterfaceName := ethernetInterface.Name
	ethernetInterfaceID := ethernetInterface.ID
	ethernetInterfaceLinkStatus := ethernetInterface.LinkStatus
	ethernetInterfaceEnabled := ethernetInterface.InterfaceEnabled
	ethernetInterfaceSpeed := fmt.Sprintf("%d Mbps", ethernetInterface.SpeedMbps)
	ethernetInterfaceState := ethernetInterface.Status.State
	ethernetInterfaceHealthState := ethernetInterface.Status.Health
	systemEthernetInterfaceLabelValues := []string{systemHostName, "ethernet_interface", ethernetInterfaceName, ethernetInterfaceID, ethernetInterfaceSpeed}
	if ethernetInterfaceStateValue, ok := parseCommonStatusState(ethernetInterfaceState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_state"].desc, prometheus.GaugeValue, ethernetInterfaceStateValue, systemEthernetInterfaceLabelValues...)

	}
	if ethernetInterfaceHealthStateValue, ok := parseCommonStatusHealth(ethernetInterfaceHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_health_state"].desc, prometheus.GaugeValue, ethernetInterfaceHealthStateValue, systemEthernetInterfaceLabelValues...)
	}
	if ethernetInterfaceLinkStatusValue, ok := parseLinkStatus(ethernetInterfaceLinkStatus); ok {

		ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_link_status"].desc, prometheus.GaugeValue, ethernetInterfaceLinkStatusValue, systemEthernetInterfaceLabelValues...)

	}

	ch <- prometheus.MustNewConstMetric(systemMetrics["system_ethernet_interface_link_enabled"].desc, prometheus.GaugeValue, boolToFloat64(ethernetInterfaceEnabled), systemEthernetInterfaceLabelValues...)

}

func parsePcieFunction(ch chan<- prometheus.Metric, systemHostName string, pcieFunction *redfish.PCIeFunction, wg *sync.WaitGroup) {

	defer wg.Done()
	pcieFunctionName := pcieFunction.Name
	pcieFunctionID := fmt.Sprint("%v", pcieFunction.ID)
	pciFunctionDeviceclass := fmt.Sprintf("%v", pcieFunction.DeviceClass)
	pciFunctionType := fmt.Sprintf("%v", pcieFunction.FunctionType)
	pciFunctionState := pcieFunction.Status.State
	pciFunctionHealthState := pcieFunction.Status.Health

	systemPCIeFunctionLabelLabelValues := []string{systemHostName, "pcie_function", pcieFunctionName, pcieFunctionID, pciFunctionDeviceclass, pciFunctionType}

	if pciFunctionStateValue, ok := parseCommonStatusState(pciFunctionState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_function_state"].desc, prometheus.GaugeValue, pciFunctionStateValue, systemPCIeFunctionLabelLabelValues...)
	}

	if pciFunctionHealthStateValue, ok := parseCommonStatusHealth(pciFunctionHealthState); ok {
		ch <- prometheus.MustNewConstMetric(systemMetrics["system_pcie_function_health_state"].desc, prometheus.GaugeValue, pciFunctionHealthStateValue, systemPCIeFunctionLabelLabelValues...)
	}
}


