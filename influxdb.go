package influxdb

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type InfluxDBClient struct {
	Host     string
	Port     int
	Database string
	Logger   *log.Logger
	client   *http.Client
}

func NewInfluxDBClient(host string, port int, database string, timeoutSec time.Duration) (InfluxDBClient, error) {

	// open file and create logger
	f, err := os.OpenFile("influxdbagent.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	l := &log.Logger{}
	l.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	l.SetOutput(io.MultiWriter(f))

	return InfluxDBClient{
		Host:     host,
		Port:     port,
		Database: database,
		Logger:   l,
		client: &http.Client{
			Timeout: timeoutSec * time.Second,
		},
	}, err
}

func (ic *InfluxDBClient) GetWriteUrl() string {
	return fmt.Sprintf("http://%v:%v/write?db=%v", ic.Host, ic.Port, ic.Database)
}

// Write string data to InfluxDB.
//
// Format of data
//
//   "<measurement>,<tag1.name>=<tag1.value>,<tag2.name>=<tag2.value>... value=<value> <time>"
//
// E.g.
//
//   "cpu_load_short,host=hacky.xyz value=0.9 1422568543703200257"
//
func (ic *InfluxDBClient) WriteString(data string) error {

	if data == "" {
		return errors.New("data is required")
	}

	r, err := http.NewRequest("POST", ic.GetWriteUrl(), strings.NewReader(data))
	if err != nil {
		return err
	}

	res, err := ic.client.Do(r)
	if err != nil {
		return err
	}

	if res.StatusCode != 204 {
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return errors.New(fmt.Sprintf("Resopnse error: %v, %v", res.StatusCode, string(resBody)))
	}
	return nil
}

func (ic *InfluxDBClient) WriteStrings(data []string) error {
	for _, s := range data {
		err := ic.WriteString(s)
		if err != nil {
			return err
		}
	}
	return nil
}
