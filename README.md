# influxdb

influxdb is InfluxDB agent for golang.

This library sends data to the InfluxDB as agent. The agent on background writes data to the InfluxDB every specified interval.


## How to use

Initialize by `infxludb.New` with several arguments.

```go
errCh := make(chan error)
agent, err = influxdb.New(host, port, database, interval)
if err != nil {
        // handle error
}
go agent.Run(errCh)
```

## Note

* Only use http protocol yet.
