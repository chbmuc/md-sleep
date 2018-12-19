package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	watchFlag    = flag.Int("w", 1000, "interval to check for i/o activity in milliseconds")
	idleFlag     = flag.Int("i", 900, "idle-time before initiating spin-down in seconds")
	rotatingFlag = flag.Bool("r", false, "spin down non-rotating slave devices (default only spin-down rotating disks)")

	hdparm  = "/usr/bin/hdparm"
	syspath = "/sys/block/"
)

// watchStats - reads the stat file at a regular intervall and return stats on change
func watchStats(inter time.Duration, device string, c chan []byte) error {
	var last []byte

	for {
		stats, err := ioutil.ReadFile(syspath + device + "/stat")
		if err != nil {
			usage(err)
		}

		if !bytes.Equal(last, stats) {
			c <- stats
			last = stats
		}

		time.Sleep(inter)
	}
}

// getSlaves - return rotating slaves of an md raid array
func getSlaves(raid string, rotational bool) ([]string, error) {
	var slaves []string

	files, err := ioutil.ReadDir(syspath + raid + "/slaves/")
	if err != nil {
		return slaves, err
	}
	for _, f := range files {
		link, err := os.Readlink(syspath + raid + "/slaves/" + f.Name())
		if err != nil {
			continue
		}

		devpath := strings.Split(link, "/")
		for i := 1; i <= 2; i++ {
			bytes, err := ioutil.ReadFile(syspath + devpath[len(devpath)-i] + "/queue/rotational")
			if err == nil {
				if rotational && bytes[0] == '1' {
					slaves = append(slaves, devpath[len(devpath)-i])
				} else if !rotational {
					slaves = append(slaves, devpath[len(devpath)-i])
				}
			}
		}
	}
	return slaves, nil
}

// spinUpDown - spin down multiple disks if up=false
//            - spin up   multiple disks if up=true
func spinUpDown(devices []string, up bool) (bool, error) {
	cmd := make([]*exec.Cmd, len(devices))

	for i, d := range devices {
		if !strings.HasPrefix(d, "/dev/") {
			d = "/dev/" + d
		}
		if up {
			cmd[i] = exec.Command(hdparm, "--read-sector", "0", d)
		} else {
			cmd[i] = exec.Command(hdparm, "-y", d)
		}
		err := cmd[i].Start()
		if err != nil {
			return !up, fmt.Errorf("hdparm for %s failed: %v", d, err)
		}
	}

	for i, d := range devices {
		err := cmd[i].Wait()
		if err != nil {
			return !up, fmt.Errorf("hdparm for %s failed: %v", d, err)
		}
	}
	return !up, nil
}

func usage(err error) {
	fmt.Println("md-sleep: watch md-raid array and spin down idle disks")
	fmt.Println()
	fmt.Println("Usage: md-sleep [-i milliseconds] [-t seconds] [-r] md-device")
	fmt.Println()
	flag.PrintDefaults()
	fmt.Println()
	if err != nil {
		fmt.Println(err)
		fmt.Println()
	}
	os.Exit(1)
}

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		usage(nil)
	}

	raid := flag.Args()[0]
	if strings.HasPrefix(raid, "/dev/") {
		raid = raid[5:]
	}

	watchInterval := time.Duration(*watchFlag) * time.Millisecond
	idleTimeout := time.Duration(*idleFlag) * time.Second

	rotating := !(*rotatingFlag)

	devices, err := getSlaves(raid, rotating)
	if err != nil {
		usage(err)
	}

	fmt.Println("Starting md-sleep for", raid, devices)
	idle, err := spinUpDown(devices, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Finished spin-up")

	blockIO := make(chan []byte, 1)
	go watchStats(watchInterval, raid, blockIO)

	for {
		if idle {
			// wait for IO
			<-blockIO
			fmt.Println("Start spin-up")
			idle, err = spinUpDown(devices, true)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println("Finished spin-up")
		} else {
			select {
			case <-blockIO:
				continue
			case <-time.After(idleTimeout):
				fmt.Println("spin-down after", idleTimeout)
				idle, err = spinUpDown(devices, false)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

	}
}
