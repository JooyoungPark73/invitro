package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eth-easl/loader/pkg/common"
	"github.com/eth-easl/loader/pkg/config"
	"github.com/eth-easl/loader/pkg/driver"
	"github.com/eth-easl/loader/pkg/trace"

	log "github.com/sirupsen/logrus"

	tracer "github.com/ease-lab/vhive/utils/tracing/go"
)

const (
	zipkinAddr = "http://localhost:9411/api/v2/spans"
)

var (
	configPath         = flag.String("config", "cmd/config_client_single.json", "Path to loader configuration file")
	verbosity          = flag.String("verbosity", "trace", "Logging verbosity - choose from [info, debug, trace]")
	iatGeneration      = flag.Bool("iatGeneration", false, "Generate iats only or run invocations as well")
	generated          = flag.Bool("generated", false, "True if iats were already generated")
	overwrite_duration = flag.Int("overwrite_duration", -1, "overwrite duration")
)

func init() {
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: time.StampMilli,
		FullTimestamp:   true,
	})
	// cfg := config.ReadConfigurationFile(*configPath)
	// logPath := fmt.Sprintf("data/logs/experiment_duration_%d.txt", cfg.ExperimentDuration)
	// file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.SetOutput(os.Stdout)

	switch *verbosity {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	cfg := config.ReadConfigurationFile(*configPath)

	if cfg.EnableZipkinTracing {
		// TODO: how not to exclude Zipkin spans here? - file a feature request
		log.Warnf("Zipkin tracing has been enabled. This will exclude Istio spans from the Zipkin traces.")

		shutdown, err := tracer.InitBasicTracer(zipkinAddr, "loader")
		if err != nil {
			log.Print(err)
		}

		defer shutdown()
	}

	if cfg.ExperimentDuration < 1 {
		log.Fatal("Runtime duration should be longer at least a minute.")
	}
	if (*overwrite_duration) > 0 {
		cfg.ExperimentDuration = *overwrite_duration
	}
	runTraceMode(&cfg, *iatGeneration, *generated)
}

func determineDurationToParse(runtimeDuration int, warmupDuration int) int {
	result := 0

	if warmupDuration > 0 {
		result += 1              // profiling
		result += warmupDuration // warmup
	}

	result += runtimeDuration // actual experiment

	return result
}

func shadowFunctions(functions []*common.Function) []*common.Function {
	newFunctions := make([]*common.Function, len(functions)*len(common.GPUSet))

	for i, f := range functions {
		for j := 0; j < len(common.GPUSet); j++ {
			copy := *f                                                     // make a copy of the function
			copy.Name = fmt.Sprintf("%s-gpu-%d", f.Name, common.GPUSet[j]) // update the name of the copy
			newFunctions[i*len(common.GPUSet)+j] = &copy                   // add the copy to the new slice
		}
		for j := 0; j < len(common.GPUSet); j++ {
			log.Infof("shadowFunctions function name is %s", f.Name)
		}
	}

	return newFunctions
}

func serverfulFunctions(functions []*common.Function) []*common.Function {
	newFunctions := make([]*common.Function, len(functions)*common.ServerfulCopyReplicas)

	for i, f := range functions {
		for j := 0; j < common.ServerfulCopyReplicas; j++ {
			copy := *f                                                 // make a copy of the function
			copy.Name = fmt.Sprintf("%s-serverful-copy-%d", f.Name, j) // update the name of the copy
			newFunctions[i*common.ServerfulCopyReplicas+j] = &copy     // add the copy to the new slice
			log.Infof("serverfulFunctions function name is %s", newFunctions[i*common.ServerfulCopyReplicas+j].Name)
		}
	}

	return newFunctions
}

func sharingFunctions(functions []*common.Function) []*common.Function {
	newFunctions := make([]*common.Function, len(functions))
	newStrings := make([]string, 0)
	var firstFuncName = ""
	for idx, f := range functions {
		parts := strings.Split(f.Name, "-")
		combined := strings.Join(parts[:2], "-")
		newStrings = append(newStrings, combined)
		if idx == 0 {
			firstFuncName = combined
		}
		fmt.Println("endingpoint :", f.Endpoint)
	}
	sharingEndpoint := functions[0].Endpoint
	// combinedName := "[" + strings.Join(newStrings, "-") + "]"
	combinedName := strings.Join(newStrings, "-")
	// copy := *functions[0]
	// copy.Name = strings.Replace(functions[0].Name, firstFuncName, combinedName, -1)
	for idx, f := range functions {
		copy := *f // make a copy of the function
		copy.UniqueName = copy.Name
		copy.Name = strings.Replace(functions[0].Name, firstFuncName, combinedName, -1) + "-" + strconv.Itoa(idx)

		copy.Endpoint = sharingEndpoint
		newFunctions[idx] = &copy // add the copy to the new slice
		log.Infof("sharing function name is %s", newFunctions[idx].Name)
	}

	return newFunctions
}

func runTraceMode(cfg *config.LoaderConfiguration, iatOnly bool, generated bool) {
	durationToParse := determineDurationToParse(cfg.ExperimentDuration, cfg.WarmupDuration)

	traceParser := trace.NewAzureParser(cfg.TracePath, durationToParse)
	functions := traceParser.Parse()
	if cfg.ContainerSharing {
		if driver.IsStringInList(cfg.ClientTraining, []string{common.Multi, common.HiveD, common.INFless, common.Elastic}) {
			functions = shadowFunctions(functions)
			functions = sharingFunctions(functions)
		}
	} else {
		if driver.IsStringInList(cfg.ClientTraining, []string{common.Multi, common.HiveD, common.INFless, common.Elastic}) {
			functions = shadowFunctions(functions)
		} else if driver.IsStringInList(cfg.ClientTraining, []string{common.Caerus, common.BatchPriority, common.PipelineBatchPriority, common.Knative}) {

		} else if driver.IsStringInList(cfg.ClientTraining, []string{common.ElasticFlow}) {
			functions = serverfulFunctions(functions)
		} else {
			log.Errorf("Invalid client_training value: %s", cfg.ClientTraining)
		}
	}

	log.Infof("Traces contain the following %d functions:\n", len(functions))
	for _, function := range functions {
		fmt.Printf("\t%s\n", function.Name)
		if len(function.Name) < 1 {
			function.UniqueName = function.Name
		}
	}

	var iatType common.IatDistribution
	switch cfg.IATDistribution {
	case "exponential":
		iatType = common.Exponential
	case "uniform":
		iatType = common.Uniform
	case "equidistant":
		iatType = common.Equidistant
	default:
		log.Fatal("Unsupported IAT distribution.")
	}

	var yamlSpecificationPath string
	switch cfg.YAMLSelector {
	case "wimpy":
		yamlSpecificationPath = "workloads/container/wimpy.yaml"
	case "container":
		yamlSpecificationPath = "workloads/container/trace_func_gpt.yaml"
	case "container-gpu":
		yamlSpecificationPath = "workloads/container/trace_func_gpt_gpu.yaml"
	case "container-dataset":
		yamlSpecificationPath = "workloads/container/trace_func_gpt_with_dataset.yaml"
	case "firecracker":
		yamlSpecificationPath = "workloads/firecracker/trace_func_go.yaml"
	default:
		log.Fatal("Invalid 'YAMLSelector' parameter.")
	}

	var traceGranularity common.TraceGranularity
	switch cfg.Granularity {
	case "minute":
		traceGranularity = common.MinuteGranularity
	case "second":
		traceGranularity = common.SecondGranularity
	default:
		log.Fatal("Invalid trace granularity parameter.")
	}

	if cfg.ClientTraining == common.INFless {
		yamlSpecificationPath = "workloads/container//trace_func_gpt_scale_constraint.yaml"
	}
	log.Infof("Using %s as a service YAML specification file.\n", yamlSpecificationPath)
	experimentDriver := driver.NewDriver(&driver.DriverConfiguration{
		LoaderConfiguration: cfg,
		IATDistribution:     iatType,
		TraceGranularity:    traceGranularity,
		TraceDuration:       durationToParse,

		YAMLPath: yamlSpecificationPath,
		TestMode: false,

		Functions: functions,
	})

	experimentDriver.RunExperiment(iatOnly, generated)
}
