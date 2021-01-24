package cgroups

import (
	"flag"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var (
	cpuExceedRate    float64
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
)

func init() {
	flag.Float64Var(&cpuExceedRate, "cpu-exceed-rate", 2.5, "Limit each cpu usages won't exceed this rate")
	flag.Uint64Var(&cpuCfsPeriod, "cpu-cfs-period", 200000, "Limit CPU CFS (Completely Fair Scheduler) period in us")
	flag.Uint64Var(&cpuCfsQuota, "cpu-cfs-quota", 200000, "Limit CPU CFS (Completely Fair Scheduler) quota in us")
	flag.Uint64Var(&cpuRtPeriod, "cpu-rt-period", 1000000, "Limit CPU Real-Time Scheduler period in us")
	flag.Uint64Var(&cpuRtRuntime, "cpu-rt-runtime", 950000, "Limit CPU Real-Time Scheduler runtime in us")
	flag.Uint64Var(&cpuShares, "cpu-shares", 1024, "CPU shares (relative weight)")
	flag.StringVar(&cpusetCpus, "cpuset-cpus", fmt.Sprintf("0-%d", runtime.NumCPU()-1), "CPUs in which to allow execution (0-3, 0,1)")
	flag.StringVar(&cpusetMems, "cpuset-mems", fmt.Sprintf("0-%d", getMemNodesNum()-1), "MEMs in which to allow execution (0-3, 0,1)")
	flag.Int64Var(&memoryLimit, "memory-limit", -1, "Memory limit in bytes; -1 indicates unlimited")
	flag.Int64Var(&memorySoftLimit, "memory-soft-limit", -1, "Memory soft limit in bytes; -1 indicates unlimited")
	flag.Int64Var(&memorySwapLimit, "memory-swap-limit", -1, "Swap limit equals to memory plus swap; -1 indicates unlimited")
	flag.Uint64Var(&memorySwappiness, "memory-swappiness", 0, "Tune container memory swappiness (range [0, 100])")
	flag.Uint64Var(&pidsMax, "pids-max", 0, "Limit pids number in container; 0 indicates unlimited")
}

func (cg *CGroups) parseCpuFlags() error {
	numCPU := uint64(runtime.NumCPU())
	if cpuExceedRate <= 0 {
		return fmt.Errorf("--cpu-exceed-rate must be positive")
	}

	if cpuCfsPeriod > 0 && cpuCfsPeriod < 1000 || cpuCfsPeriod > 1000000 {
		return fmt.Errorf("--cpu-cfs-period requires [1000, 1000000]")
	}
	cg.cpuCfsPeriod = cpuCfsPeriod

	if cpuCfsQuota > uint64(float64(cpuCfsPeriod*numCPU)*cpuExceedRate) {
		return fmt.Errorf("--cpu-cfs-quota can't exceed cpuCfsPeriod*numCPU*cpuExceedRate")
	}
	cg.cpuCfsQuota = cpuCfsQuota

	if cpuRtPeriod > 2000000 {
		return fmt.Errorf("--cpu-rt-period can't exceed 2000000")
	}
	cg.cpuRtPeriod = cpuRtPeriod

	if cpuRtRuntime > uint64(float64(cpuRtPeriod*numCPU)*cpuExceedRate) {
		return fmt.Errorf("--cpu-rt-runtime can't exceed cpuRtPeriod*numCPU*cpuExceedRate")
	}
	cg.cpuRtRuntime = cpuRtRuntime

	if cpuShares > 0 && cpuShares < 2 {
		return fmt.Errorf("--cpu-shares requires >= 2")
	}
	cg.cpuShares = cpuShares

	return nil
}

func (cg *CGroups) parseCpusetFlags() error {
	numCPU := runtime.NumCPU()
	numMem := getMemNodesNum()

	if err := validateCpusetArgs(cpusetCpus, "cpu", numCPU); err != nil {
		return err
	}
	cg.cpusetCpus = cpusetCpus

	if err := validateCpusetArgs(cpusetMems, "mem", numMem); err != nil {
		return err
	}
	cg.cpusetMems = cpusetMems

	return nil
}

func validateCpusetArgs(args, cls string, maxNum int) error {
	if args == "" {
		return nil
	}

	err := fmt.Errorf("--cpuset-%ss requires a-b, "+
		"and a must be less or equal to b, and both "+
		"must be in the range [0, %d)", cls, maxNum)

	for _, arg := range strings.Split(args, ",") {
		if strings.Contains(arg, "-") {
			re, _ := regexp.Compile(`(\d+)-(\d+)`)
			if !re.MatchString(args) {
				return err
			}

			pairs := re.FindStringSubmatch(arg)
			a, _ := strconv.Atoi(pairs[1])
			b, _ := strconv.Atoi(pairs[2])

			if a >= 0 && b < maxNum && a <= b {
				continue
			}
			return err
		}

		numArg, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("%s is not a number", arg)
		}
		if numArg < 0 || numArg >= maxNum {
			return fmt.Errorf("value of --cpuset-%ss must "+
				"be in the range [0, %d)", cls, maxNum)
		}
	}

	return nil
}

func (cg *CGroups) parseMemoryFlags() error {
	if memoryLimit < 0 {
		memoryLimit = -1
	}
	cg.memoryLimit = memoryLimit

	if memorySoftLimit < 0 {
		memorySoftLimit = -1
	}
	if memorySoftLimit > -1 && memoryLimit > -1 && memorySoftLimit < memoryLimit {
		return fmt.Errorf("memorySoftLimit requires >= memoryLimit")
	}
	cg.memorySoftLimit = memorySoftLimit

	if memorySwapLimit < 0 {
		memorySwapLimit = -1
	}
	if memorySwapLimit > -1 && memoryLimit > -1 && memorySwapLimit < memoryLimit {
		return fmt.Errorf("memorySwapLimit requires >= memoryLimit")
	}
	cg.memorySwapLimit = memorySwapLimit

	if memorySwappiness > 100 {
		memorySwappiness = 100
	}
	cg.memorySwappiness = memorySwappiness

	return nil
}
