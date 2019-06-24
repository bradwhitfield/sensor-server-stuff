package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// $env:AUTH_TOKEN="ciPWXJf9abMEr83tnZlBBO7Ad7VQyQ_YK6f8h7vxAfqmjv0CfMzhLihXGrSudt6ZN64lbzGs12qsa4FEShsuzQ=="
// $env:ORG="sensor"
// $env:BUCKET="data"
// $env:INFLUX_HOST="localhost:9999"

// InfluxDB represents the fields required to leverage InfluxDB
type InfluxDB struct {
	AuthToken string
	Org       string
	Bucket    string
	Host      string
	Client    *http.Client
}

// Write loads data into InfluxDB
func (db *InfluxDB) Write(measuerment string, tags map[string]string, fieldKey string, fieldValue string) error {
	// assemble the data in line protocol format
	// Reference at https://v2.docs.influxdata.com/v2.0/reference/line-protocol/
	var data bytes.Buffer
	data.WriteString(measuerment)
	for k, v := range tags {
		data.WriteString(",")
		data.WriteString(k)
		data.WriteString("=")
		data.WriteString(v)
	}
	data.WriteString(" ")
	data.WriteString(fieldKey)
	data.WriteString("=")
	data.WriteString(fieldValue)
	data.WriteString(" ")
	data.WriteString(strconv.FormatInt(time.Now().Unix(), 10))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v2/write?org=%s&bucket=%s&precision=s", db.Host, db.Org, db.Bucket), bytes.NewReader(data.Bytes()))
	req.Header = http.Header{
		"Authorization": []string{fmt.Sprintf("Token %s", db.AuthToken)},
	}
	resp, err := db.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed performing POST to Influx")
	}

	if resp.StatusCode != http.StatusNoContent {
		if resp.Body != nil {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return errors.Wrap(err, "reading the body failed")
			}
			return fmt.Errorf("Influx did not return 204 as expected - status: %d, body: %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("Influx did not return 204 as expected - status: %d", resp.StatusCode)
	}

	return nil
}
