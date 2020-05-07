package internal

import (
	"encoding/xml"
	"github.com/jstemmer/go-junit-report/formatter"
	"github.com/jtaczanowski/go-graphite-client"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

func ReadJunitReport(report io.Reader) (*formatter.JUnitTestSuites, error) {
	dat, err := ioutil.ReadAll(report)
	if err != nil {
		return nil, err
	}
	suites := &formatter.JUnitTestSuites{}
	err = xml.Unmarshal(dat, suites)
	if err != nil {
		return nil, err
	}

	return suites, nil
}

func MapToGraphiteFormat(suites *formatter.JUnitTestSuites) (map[string]float64, error) {
	metricsMap := map[string]float64{}
	for _, suit := range suites.Suites {
		if len(suit.TestCases) == 0 {
			continue
		}
		for _, testcase := range suit.TestCases {
			testName := ""
			testStatus := ""

			if testcase.Classname != "" {
				testName = testcase.Classname
			}
			if testcase.Name != "" {
				testName = testName + "." + testcase.Name
			}
			testName = strings.ReplaceAll(testName, " ", "")
			testName = strings.ReplaceAll(testName, "\"", "")
			testName = strings.ReplaceAll(testName, ":", ".")
			testName = strings.ReplaceAll(testName, "|", ".")
			testName = strings.ReplaceAll(testName, ",", "_")
			testName = strings.ReplaceAll(testName, "/", ".")
			testName = strings.ReplaceAll(testName, "{", "")
			testName = strings.ReplaceAll(testName, "}", "")

			testStatus = "OK"
			if testcase.Failure != nil {
				testStatus = "Error"
			}
			if testcase.SkipMessage != nil {
				testStatus = "Skipped"
			}

			testValue, err := strconv.ParseFloat(testcase.Time, 64)
			if err != nil {
				return nil, err
			}

			metric := testName + "." + testStatus

			metricsMap[metric] = testValue
		}
	}

	return metricsMap, nil
}

func SendToGraphite(host string, port int, prefix string, metrics map[string]float64) error {
	graphiteClient := graphite.NewClient(host, port, prefix, "tcp")
	err := graphiteClient.SendData(metrics)
	if err != nil {
		return err
	}
	return nil
}