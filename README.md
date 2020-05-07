# go-graphite-report 

go-graphite-report - simple Golang tool, reads test report file in junit xml format and sends it in batch 
to graphite server. 


## Installation

Go version 1.2 or higher is required. Install or update using the `go get`
command:

```bash
go get -u github.com/lyubenkov/go-graphite-report
```

## Usage
```bash
go-graphite-report -f ./junit-report.xml -h mygraphiteserver.com -p 1234 -x myprefix
```
