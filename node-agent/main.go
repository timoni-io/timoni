package main

import (
	"bytes"
	"encoding/json"
	"io"
	"lib/tlog"
	"lib/utils"
	"lib/utils/maps"
	"lib/utils/slice"
	"lib/utils/types"
	"math"
	"metric-node-agent/global"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type ProcessData struct {
	P          *process.Process
	TotalTime0 float64
	CputTotal0 float64
	Cgroup     string
}

type metricDataS struct {
	Metric string            `json:"metric"`
	Value  float64           `json:"value"`
	Tags   map[string]string `json:"tags"`
}

type podCacheData struct {
	EnvID       string `json:"EnvID"`
	ElementName string `json:"ElementName"`
}

type podMetric struct {
	CPU         MagicSlice
	RSS         MagicSlice
	ElementName string
}

type tmpMetric struct {
	CPU         float64
	RSS         float64
	ElementName string
}

type MagicSlice struct {
	*slice.Rigid[int32, uint16]
}

const (
	emptyPodID = "-"
)

type podMetricType struct {
	*maps.SafeMap[string, *podMetric]
}

const timeRange uint16 = 120

var (
	podCacheMap   = maps.NewSafe[string, podCacheData](nil)   // podName -> podCacheData
	podMetricsMap = maps.NewSafe[string, *podMetricType](nil) // envName -> podID -> podMetric
)

func main() {
	nodeIP := os.Getenv("NODE_IP")
	if global.VMAgentAddr == "" {
		tlog.Fatal("env VM_AGENT_ADDR is empty")
	}
	tlog.Info("starting {{GitTag}}...", tlog.Vars{
		"GitTag": global.GitTag,
	})

	updatePodInfo()

	// For elastic search, run:
	exec.Command("/sbin/sysctl", "-w", "vm.max_map_count=262144").Run()

	for {
		processData := initProcesses()
		var md []metricDataS
		tmp := maps.NewSafe[string, *maps.SafeMap[string, *tmpMetric]](nil)
		for _, pd := range processData {

			// ---

			processName, err := pd.P.Name()
			if err != nil {
				continue
			}

			if !strings.Contains(pd.Cgroup, "kubepods") {
				continue
			}

			podID := emptyPodID

			// cgroup v2
			for _, str := range strings.Split(pd.Cgroup, "/") {
				if strings.Contains(str, "kubepods") {
					// tlog.PrintJSON(pd)
					_, after, found := strings.Cut(str, "-pod")
					if found {
						if len(after) > 0 {
							podID = after[:len(after)-6]
						}
					}
				}
			}

			if podID == emptyPodID {
				// cgroup v1
				for _, str := range strings.Split(pd.Cgroup, "/") {
					if strings.HasPrefix(str, "pod") {
						podID = str[3:]
					}
				}
			}

			if podID == emptyPodID {
				tlog.Warning("podID is empty", tlog.Vars{
					"pd.Cgroup": pd.Cgroup,
					"podID":     podID,
				})

			} else {
				podID = strings.ReplaceAll(podID, "-", "_")
			}

			if processName == "pause" && podID != "" {
				// skip pause process
				continue
			}

			// ---

			cpuCurrent, err := pd.CPUCurrentUtilization()
			if err != nil {
				continue
			}
			cpuCurrent = float64(math.Round(cpuCurrent*10) / 10)

			memInfo, err := pd.P.MemoryInfo()
			if tlog.Error(err) != nil {
				continue
			}

			// ---------------------------------
			podData := podCacheMap.Get(podID)

			if podData.EnvID == "" {
				tlog.Warning("podData.EnvID is empty", tlog.Vars{
					"podID": podID,
				})
				tlog.PrintJSON(podCacheMap)
			}

			tags := map[string]string{
				"name": processName,
				// "pod_id":  podID,
				"node_ip": nodeIP,
				"env_id":  podData.EnvID,
				"element": podData.ElementName,
			}
			for k, v := range tags {
				if v == "" {
					tags[k] = "-"
				}
			}

			if podData.EnvID != "" && podData.ElementName != "" {
				envMap := tmp.Get(podData.EnvID)
				if envMap == nil {
					envMap = maps.NewSafe[string, *tmpMetric](nil)
					tmp.Set(podData.EnvID, envMap)
				}

				podMap := envMap.Get(podID)
				if podMap == nil {
					podMap = &tmpMetric{ElementName: podData.ElementName}
					envMap.Set(podID, podMap)
				}
				podMap.CPU += cpuCurrent
				podMap.RSS += float64(memInfo.RSS / 1024) // KiB
			}

			md = append(md,
				metricDataS{
					Metric: "timoni_process_cpu_utilization",
					Value:  cpuCurrent,
					Tags:   tags,
				},
				metricDataS{
					Metric: "timoni_process_rss_utilization",
					Value:  float64(memInfo.RSS / 1024), // KiB
					Tags:   tags,
				},
			)
		}
		go updatePodMetricMap(tmp.Iter())
		go send(md)

		time.Sleep(1 * time.Second)
	}
}

func updatePodMetricMap(iter types.Iterator[string, *maps.SafeMap[string, *tmpMetric]]) {
	for v := range iter {
		podMap := podMetricsMap.Get(v.Key)
		if podMap == nil {
			podMap = &podMetricType{maps.NewSafe[string, *podMetric](nil)}
			podMetricsMap.Set(v.Key, podMap)
		}
		for vv := range v.Value.Iter() {
			podMetr := podMap.Get(vv.Key)
			if podMetr == nil {
				podMetr = &podMetric{
					CPU:         MagicSlice{slice.NewRigid[int32](timeRange).Safe()},
					RSS:         MagicSlice{slice.NewRigid[int32](timeRange).Safe()},
					ElementName: vv.Value.ElementName,
				}
				podMap.Set(vv.Key, podMetr)
			}
			podMetr.CPU.Add(int32(math.Floor(vv.Value.CPU)))
			podMetr.RSS.Add(int32(vv.Value.RSS))

		}
	}
	updatePodInfo()
}

func initProcesses() []ProcessData {
	processes, errP := process.Processes()
	tlog.Fatal(errP)
	var processData []ProcessData
	pidsCgroupsMap := getPidsCgroupsMap()

	for _, p := range processes {
		crt_time, errCT := p.CreateTime()
		// tlog.Error(errCT)
		cput, errT := p.Times()
		// tlog.Error(errT)
		if errCT == nil && errT == nil {
			processData = append(processData, ProcessData{p, time.Since(time.Unix(0, crt_time*int64(time.Millisecond))).Seconds(), cput.Total(), pidsCgroupsMap[p.Pid]})
		}
	}

	return processData
}

func send(md []metricDataS) {
	// tlog.PrintJSON(md)
	buf, err := json.Marshal(md)
	if tlog.Error(err) != nil {
		return
	}
	// fmt.Println("http://"+config.VMAgentAddr+"/api/put")
	// fmt.Println(string(buf))
	// fmt.Println("---")
	// return
	res, err := http.Post(
		"http://"+global.VMAgentAddr+"/api/put",
		"application/json",
		bytes.NewReader(buf),
	)
	if tlog.Error(err) != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 204 {
		buf, _ = io.ReadAll(res.Body)
		tlog.Error("Wystąpił błąd przy insercie do VM. \n" + string(buf))
	}
}

func (pd *ProcessData) CPUCurrentUtilization() (float64, error) {
	crt_time, errCT := pd.P.CreateTime()
	if errCT != nil {
		return 0.0, errCT
	}
	cput, errT := pd.P.Times()
	if errT != nil {
		return 0.0, errT
	}
	totalTime := time.Since(time.Unix(0, crt_time*int64(time.Millisecond))).Seconds()
	// fmt.Println((cput.Total() - pd.CputTotal0) / (totalTime - pd.TotalTime0) * 100)
	return 100 * (cput.Total() - pd.CputTotal0) / (totalTime - pd.TotalTime0), nil
}

func getPidsCgroupsMap() map[int32]string {
	cmd := []string{"ps", "xawf", "-eo", "pid,cgroup"}
	out, errC := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if errC != nil {
		tlog.Error(errC, tlog.Vars{
			"out": string(out),
			"cmd": strings.Join(cmd, " "),
		})
		return nil
	}
	pidsCgroups := strings.Split(string(out), "\n")[1:]
	pidsCgroups = pidsCgroups[:len(pidsCgroups)-1]
	pidsCgroupsMap := make(map[int32]string)

	for _, pc := range pidsCgroups {
		pcs := strings.Split(strings.TrimSpace(pc), " ")
		if pcs[1] != "-" {
			pid, errA := strconv.Atoi(pcs[0])
			tlog.Fatal(errA)
			pidsCgroupsMap[int32(pid)] = pcs[1]
		}
	}

	return pidsCgroupsMap
}

func updatePodInfo() {
	// tlog.PrintJSON(podMetricsMap)
	res, err := http.Post(global.TimoniURL+"/api/system-pod-cache", "application/json", bytes.NewBuffer(utils.Must(json.Marshal(podMetricsMap))))
	if tlog.Error(err) != nil {
		return
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if tlog.Error(err) != nil {
		return
	}

	if buf[0] != 1 {
		tlog.Error("api return error: " + string(buf[1]))
		return
	}

	tmp := map[string]podCacheData{}
	err = json.Unmarshal(buf[1:], &tmp)
	if tlog.Error(err) != nil {
		return
	}
	// tlog.PrintJSON(tmp)
	podCacheMap = maps.New(tmp).Safe()
	go deleteOldMetrics()
}

func deleteOldMetrics() {
	// tlog.PrintJSON(podCacheMap)
	for env := range podMetricsMap.Iter() {
		for pod := range env.Value.Iter() {
			// if pod.Value.ElementName == "cpu-spammer" {
			// 	fmt.Println("check pod: " + pod.Key + " elementName: " + pod.Value.ElementName)
			// 	fmt.Println(podCacheMap.GetFull(pod.Key))
			// }
			if _, ok := podCacheMap.GetFull(pod.Key); !ok {
				// fmt.Println("drop pod: " + pod.Key + " elementName: " + pod.Value.ElementName)
				podMetricsMap.Get(env.Key).Delete(pod.Key)
			}
		}
		if env.Value.Len() == 0 {
			podMetricsMap.Delete(env.Key)
		}
	}
}

func (m *MagicSlice) MarshalJSON() ([]byte, error) {
	// we calculate the average value of the metric
	// for the last 15 minutes

	var avg float64 = 0
	for _, c := range m.GetAll() {
		avg += float64(c)
	}
	avg /= float64(m.Len())

	return json.Marshal(avg)
}

func (m *podMetricType) MarshalJSON() ([]byte, error) {
	tmp := map[string]*podMetric{}
	for pod := range m.Iter() {
		tmp[pod.Value.ElementName] = pod.Value
	}
	return json.Marshal(tmp)
}
