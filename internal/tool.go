package internal

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/jstemmer/go-junit-report/formatter"
	"github.com/jtaczanowski/go-graphite-client"
	"github.com/rainycape/unidecode"
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
		suiteName := "suite"
		if suit.Name != "" {
			suiteName = suit.Name
		}
		for _, testcase := range suit.TestCases {
			testName := "testcase"

			if !strings.Contains(testcase.Name, suiteName) {
				suiteName = testcase.Classname
			}
			if testcase.Name != "" {
				testName = testcase.Name
			}

			// sometimes there's a json after |
			r := strings.Split(testName, "|")
			if len(r) > 1 {
				r = r[1:]
				s := strings.Join(r, "")
				var j map[string]string
				err := json.Unmarshal([]byte(s), &j)
				if err == nil {
					buf := bytes.Buffer{}
					for key, value := range j {
						buf.WriteString(key + "_" + value)
					}
					testName = buf.String()
				} else {
					testName = s
				}
			}

			testStatus := "OK"
			if testcase.Failure != nil {
				testStatus = "Error"
			}
			if testcase.SkipMessage != nil {
				testStatus = "Skipped"
			}

			metric := strings.Join([]string{filterString(suiteName), filterString(testName), testStatus}, ".")
			metric = unidecode.Unidecode(metric)

			// remove adjacent duplicates of _
			var last rune
			var buf strings.Builder
			for _, r := range metric {
				if r != '_' {
					buf.WriteRune(r)
					last = -1
				} else if r != last {
					buf.WriteRune(r)
					last = r
				}
			}
			metric = buf.String()

			testValue, err := strconv.ParseFloat(testcase.Time, 64)
			if err != nil {
				return nil, err
			}

			metricsMap[metric] = testValue
		}
	}

	return metricsMap, nil
}

func filterString(s string) string {
	r := s
	r = strings.ReplaceAll(r, " ", "_")
	r = strings.ReplaceAll(r, ".", "_")
	r = strings.ReplaceAll(r, "-", "_")
	r = strings.ReplaceAll(r, "\"", "")
	r = strings.ReplaceAll(r, ":", "_")
	r = strings.ReplaceAll(r, "|", "_")
	r = strings.ReplaceAll(r, ",", "_")
	r = strings.ReplaceAll(r, "/", "_")
	r = strings.ReplaceAll(r, "{", "")
	r = strings.ReplaceAll(r, "}", "")
	r = strings.ReplaceAll(r, "'", "")

	return r
}

func SendToGraphite(host string, port int, prefix string, metrics map[string]float64) error {
	graphiteClient := graphite.NewClient(host, port, prefix, "tcp")
	err := graphiteClient.SendData(metrics)
	if err != nil {
		return err
	}

	return nil
}