package influxdb

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

func New(host string, port int, database string, postInterval time.Duration) (*MemAgent, error) {

	// check setting value existence
	if host == "" || port <= 0 || database == "" {
		return nil, errors.New(fmt.Sprintf(
			"Invalid config: host=%s, port=%v, database=%s", host, port, database))
	}

	client, err := NewInfluxDBClient(host, port, database, 10)
	if err != nil {
		return nil, err
	}
	client.Logger.Printf("Created InfluxDB Client: host=%v, port=%v, database=%v\n", host, port, database)

	return &MemAgent{
		client:   client,
		lock:     &sync.Mutex{},
		interval: postInterval,
	}, nil
}

func (agent *MemAgent) Add(record string) {
	agent.records = append(agent.records, record)
}

// Add record with each value
func (agent *MemAgent) AddWith(measurement string, tag string, value float64) {
	agent.records = append(agent.records, fmt.Sprintf("%s,%s value=%v %v", measurement, tag, value, time.Now().UnixNano()))
}

func (agent *MemAgent) Adds(records []string) {
	for _, record := range records {
		agent.records = append(agent.records, record)
	}
}

func (agent *MemAgent) Run(errCh chan<- error) {
	for {
		time.Sleep(agent.interval)

		// write records to InfluxDB if there is records
		if len(agent.records) > 0 {
			to := len(agent.records)
			start := time.Now()
			for i := 0; i < to; i++ {
				err := agent.client.WriteString(agent.records[i])
				if err != nil {
					agent.client.Logger.Println("Failed to write record.", err)
					errCh <- err
				}
			}

			// remove the records written
			agent.records = agent.records[to:]
			agent.client.Logger.Printf("[agent] Finished write. time=%v, len=%v\n", time.Now().Sub(start), to)
		}
	}
}
