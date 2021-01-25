package cgroups

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	procsFilename        = "cgroup.procs"
	memoryName           = "memory"
	memoryLimitName      = "memory.limit_in_bytes"
	memorySoftLimitName  = "memory.soft_limit_in_bytes"
	memorySwapLimitName  = "memory.memsw.limit_in_bytes"
	memorySwappinessName = "memory.swappiness"

	cpuName          = "cpu,cpuacct"
	cpuCfsPeriodName = "cpu.cfs_period_us"
	cpuCfsQuotaName  = "cpu.cfs_quota_us"
	cpuRtPeriodName  = "cpu.rt_period_us"
	cpuRtRuntimeName = "cpu.rt_runtime_us"
	cpuSharesName    = "cpu.shares"

	cpusetName     = "cpuset"
	cpusetCpusName = "cpuset.cpus"
	cpusetMemsName = "cpuset.mems"

	pidsName    = "pids"
	pidsMaxName = "pids.max"

	cgroupName     = "cgroup"
	cgroupInfoFile = "/proc/self/cgroup"
	mountInfoFile  = "/proc/self/mountinfo"
)

// CGroups ...
type CGroups struct {
	Path string
	Pid  int

	cpuCfsPeriod     uint64
	cpuCfsQuota      uint64
	cpuRtPeriod      uint64
	cpuRtRuntime     uint64
	cpuShares        uint64
	cpusetCpus       string
	cpusetMems       string
	memoryLimit      int64
	memorySoftLimit  int64
	memorySwapLimit  int64
	memorySwappiness uint64
	pidsMax          uint64
}

// NewCGroup creates an empty CGroups
func NewCGroup(path string) *CGroups {
	if len(path) == 0 {
		panic("Invalid cgroup path")
	}
	return &CGroups{Path: path}
}

func (cg *CGroups) setCPU() error {
	if !subsystemIsMounted(cpuName) {
		return fmt.Errorf("subsystem %s is not mounted", cpuName)
	}

	// cpuPath: /sys/fs/cgroup/cpu/<cg.Path>/
	cpuPath, err := getSubsystemPath(cpuName, cg.Path)
	if err != nil {
		return err
	}

	var cpuFile string
	var value []byte

	cpuFile = path.Join(cpuPath, cpuRtPeriodName)
	value = []byte(strconv.FormatUint(cg.cpuRtPeriod, 10))
	if err := writeFile(cpuFile, value); err != nil {
		return err
	}

	cpuFile = path.Join(cpuPath, cpuRtRuntimeName)
	value = []byte(strconv.FormatUint(cg.cpuRtRuntime, 10))
	if err := writeFile(cpuFile, value); err != nil {
		return err
	}

	cpuFile = path.Join(cpuPath, cpuSharesName)
	value = []byte(strconv.FormatUint(cg.cpuShares, 10))
	if err := writeFile(cpuFile, value); err != nil {
		return err
	}

	cpuFile = path.Join(cpuPath, cpuCfsPeriodName)
	value = []byte(strconv.FormatUint(cg.cpuCfsPeriod, 10))
	if err := writeFile(cpuFile, value); err != nil {
		return err
	}

	cpuFile = path.Join(cpuPath, cpuCfsQuotaName)
	value = []byte(strconv.FormatUint(cg.cpuCfsQuota, 10))
	if err := writeFile(cpuFile, value); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	return nil
}

func (cg *CGroups) setCpuset() error {
	if !subsystemIsMounted(cpusetName) {
		return fmt.Errorf("subsystem %s is not mounted", cpusetName)
	}

	// cpusetPath: /sys/fs/cgroup/cpuset/<cg.Path>/
	cpusetPath, err := getSubsystemPath(cpusetName, cg.Path)
	if err != nil {
		return err
	}

	var cpusetFile string
	var value []byte

	cpusetFile = path.Join(cpusetPath, cpusetCpusName)
	value = []byte(cg.cpusetCpus)
	if err := writeFile(cpusetFile, value); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	cpusetFile = path.Join(cpusetPath, cpusetMemsName)
	value = []byte(cg.cpusetMems)
	if err := writeFile(cpusetFile, value); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	return nil
}

func (cg *CGroups) setMemory() error {
	if !subsystemIsMounted(memoryName) {
		return fmt.Errorf("subsystem %s is not mounted", memoryName)
	}

	// memoryPath: /sys/fs/cgroup/memory/<cg.Path>/
	memoryPath, err := getSubsystemPath(memoryName, cg.Path)
	if err != nil {
		return err
	}

	var memoryFile string
	var value []byte

	memoryFile = path.Join(memoryPath, memoryLimitName)
	value = []byte(strconv.FormatInt(cg.memoryLimit, 10))
	if err := writeFile(memoryFile, value); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	memoryFile = path.Join(memoryPath, memorySoftLimitName)
	value = []byte(strconv.FormatInt(cg.memorySoftLimit, 10))
	if err := writeFile(memoryFile, value); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	memoryFile = path.Join(memoryPath, memorySwapLimitName)
	value = []byte(strconv.FormatInt(cg.memorySwapLimit, 10))
	if err := writeFile(memoryFile, value); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	memoryFile = path.Join(memoryPath, memorySwappinessName)
	value = []byte(strconv.FormatUint(cg.memorySwappiness, 10))
	if err := writeFile(memoryFile, value); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	return nil
}

func (cg *CGroups) setPids() error {
	if !subsystemIsMounted(pidsName) {
		return fmt.Errorf("subsystem %s is not mounted", pidsName)
	}

	// pidsPath: /sys/fs/cgroup/pids/<cg.Path>/
	pidsPath, err := getSubsystemPath(pidsName, cg.Path)
	if err != nil {
		return err
	}

	var pidsFile string
	var value []byte

	pidsFile = path.Join(pidsPath, pidsMaxName)
	value = []byte(strconv.FormatUint(cg.pidsMax, 10))
	if err := writeFile(pidsFile, value); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	return nil
}

func (cg *CGroups) setProcs() error {
	// procsPath: /sys/fs/cgroup/memory/<cg.Path>/
	procsPath, err := getSubsystemPath(memoryName, cg.Path)
	if err != nil {
		return err
	}

	var procsFile string
	var value []byte

	procsFile = path.Join(procsPath, procsFilename)
	value = []byte(strconv.Itoa(cg.Pid))
	if err := writeFile(procsFile, value); err != nil {
		fmt.Printf("err %v\n", err)
		return err
	}

	return nil
}

//func fileOrDirExists(fileOrDir string) (bool, error) {
//	if _, err := os.Stat(fileOrDir); err != nil {
//		if os.IsNotExist(err) {
//			return false
//		}
//	}
//	return true
//}

func writeFile(filename string, value []byte) (err error) {
	err = ioutil.WriteFile(filename, value, 0644)
	if err != nil {
		fmt.Printf("write %s, value %s, err %v\n", filename, string(value), err)
	}
	return err
}

func subsystemIsMounted(subsystemRootName string) bool {
	contentsBytes, err := ioutil.ReadFile(cgroupInfoFile)
	if err != nil {
		return false
	}

	for _, subsystemInfo := range strings.Split(string(contentsBytes), "\n") {
		subFields := strings.Split(subsystemInfo, ":")
		if len(subFields) < 2 {
			continue
		}
		if subFields[1] == subsystemRootName ||
			(subFields[1] == "cpuacct,cpu" &&
				subsystemRootName == cpuName) {
			return true
		}
	}

	return false
}

func getSubsystemPath(subsystemRootName, cgPath string) (string, error) {
	if cgPath == "" {
		return "", fmt.Errorf("Invalid cgPath")
	}
	rootMntPoint, err := getSubsystemMountPoint(subsystemRootName)
	if err != nil {
		return "", fmt.Errorf("failed to get root mountpoint of %s: %v",
			subsystemRootName, err)
	}

	subsystemPath := path.Join(rootMntPoint, cgPath)
	// ensure the subsystemPath always exists.
	if err := os.MkdirAll(subsystemPath, 0755); err != nil {
		return "", fmt.Errorf("failed to mkdir %s: %v",
			subsystemPath, err)
	}

	return subsystemPath, nil
}

// subsystemRootName = cpu, return /sys/fs/cgroup
func getSubsystemMountPoint(subsystemRootName string) (string, error) {
	contentsBytes, err := ioutil.ReadFile(mountInfoFile)
	if err != nil {
		return "", err
	}

	for _, mntInfo := range strings.Split(string(contentsBytes), "\n") {
		mntFields := strings.Split(mntInfo, " ")
		if len(mntFields) < 8 {
			continue
		}
		if mntFields[7] == cgroupName && mntFields[8] == cgroupName {
			if strings.HasSuffix(mntFields[4], subsystemRootName) {
				return mntFields[4], nil
			}
		}
	}

	return "", fmt.Errorf("subsystem %s not mounted", subsystemRootName)
}

// get value from /sys/fs/cgroup/cpuset/cpuset.mems
// or use the command: `numactl --hardware`
// this function ignores errors on purpose.
func getMemNodesNum() int {
	cpusetRoot, _ := getSubsystemMountPoint(cpusetName)
	confFile := path.Join(cpusetRoot, cpusetMemsName)

	valueBytes, _ := ioutil.ReadFile(confFile)
	value := string(valueBytes[:len(valueBytes)-1])

	re, _ := regexp.Compile(`^[\d,-]*(\d+)$`)
	results := re.FindStringSubmatch(value)

	memNodesNum, _ := strconv.Atoi(results[1])
	return memNodesNum + 1
}
