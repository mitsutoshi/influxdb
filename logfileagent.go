package influxdb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	newLineLiteral = "\\n"
	label          = "INFLUX:"
)

func NewLogfileAgent(host string, port int, database string, postInterval time.Duration) (*LogfileAgent, error) {

	if host == "" || port <= 0 || database == "" {
		return nil, errors.New(fmt.Sprintf(
			"Invalid config: host=%s, port=%v, database=%s", host, port, database))
	}

	client, err := NewInfluxDBClient(host, port, database, 10)
	if err != nil {
		return nil, err
	}
	client.Logger.Printf("Created InfluxDB Client: host=%v, port=%v, database=%v\n", host, port, database)

	return &LogfileAgent{
		client:   client,
		lock:     &sync.Mutex{},
		interval: postInterval,
	}, nil
}

func (agent *LogfileAgent) Run(name string, errCh chan<- error) {

	fp, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	posFileName := name + ".pos"
	offset, _ := getPos(posFileName)

	var prevOffset int64 = -1
	for {
		time.Sleep(agent.interval * time.Second)

		if offset == prevOffset {
			continue
		}

		_, err = fp.Seek(offset, 0)
		if err != nil {
			errCh <- err
		}
		prevOffset = offset

		b, err := ioutil.ReadAll(fp)
		if err != nil {
			errCh <- err
		}
		if len(b) == 0 {
			continue
		}

		offset += int64(len(b))

		err = writePos(name+".pos", offset)
		if err != nil {
			errCh <- err
		}
		fmt.Printf("readsize:%v, nextoffset: %v\n", len(b), offset)

		lines := strings.Split(string(b), "\n")
		records := parse(lines)

		for _, s := range records {
			err := agent.client.WriteString(s)
			if err != nil {
				log.Println("Failed to write record.", err)
			}
		}
	}
}

func parse(lines []string) []string {
	var records []string
	var record string
	for _, line := range lines {
		if strings.Contains(line, label) {
			tokens := strings.Split(line, label)
			if strings.Contains(tokens[1], newLineLiteral) {
				record = strings.Replace(tokens[1], newLineLiteral, "\n", 99)
			} else {
				record = tokens[1]
			}
			records = append(records, record)
		}
	}
	return records
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func getPos(posFileName string) (int64, error) {
	if exists(posFileName) {
		posFile, err := os.Open(posFileName)
		if err != nil {
			return 0, err
		}

		pos, _ := ioutil.ReadAll(posFile)
		posFile.Close()

		if len(pos) != 0 {
			log.Println("Resume logfile pos:", string(pos))
			return strconv.ParseInt(string(pos), 10, 64)
		}
	}
	return 0, nil
}

func writePos(name string, offset int64) error {
	posFile, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer posFile.Close()
	_, err = posFile.WriteString(strconv.FormatInt(offset, 10))
	if err != nil {
		return err
	}
	log.Println("Save pos", offset)
	return nil
}
