package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/loafoe/mmdb"
	"github.com/oschwald/geoip2-golang"
)

// Config
type Config struct {
	LicenseKey string `json:"license_key"`
}

//nolint
func New() interface{} {
	return &Config{}
}

var dbMutex sync.Mutex

//nolint
func initDB(licenseKey string) (*geoip2.Reader, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	file, err := ioutil.TempFile("", "*.mmdb")
	if err != nil {
		return nil, err
	}
	resp, err := mmdb.Download(file.Name(), licenseKey)
	if err != nil {
		err = fmt.Errorf("licenseKey: '%s', resp: %v, error: %w", licenseKey, resp, err)
		return nil, err
	}
	reader, err := geoip2.Open(file.Name())
	if err != nil {
		return nil, err
	}
	return reader, nil
}

var db *geoip2.Reader

var doOnce sync.Once
var dbErr error

// Access implements the Access step
func (conf Config) Access(kong *pdk.PDK) {

	doOnce.Do(func() {
		db, dbErr = initDB(conf.LicenseKey)
	})

	if db == nil {
		_ = kong.ServiceRequest.SetHeader("X-Detected-Country-Error", fmt.Sprintf("GeoIP database not ready: %v", dbErr))
		headers := map[string][]string{
			"X-Kong-Geofence": {"active"},
		}
		kong.Response.Exit(http.StatusForbidden, fmt.Sprintf("GeoIP database not ready: %v", dbErr), headers)
		return
	}
	clientIP, err := kong.Client.GetIp()
	if err != nil {
		_ = kong.ServiceRequest.SetHeader("X-Detected-Country-Error", err.Error())
		return
	}
	_ = kong.ServiceRequest.SetHeader("X-Detected-IP", clientIP)
	ip := net.ParseIP(clientIP)
	record, err := db.Country(ip)
	if err != nil {
		_ = kong.ServiceRequest.SetHeader("X-Detected-Country-Error", err.Error())
		return
	}
	countryHeader := fmt.Sprintf("[%v]", record.Country.IsoCode)
	_ = kong.ServiceRequest.SetHeader("X-Detected-Country", countryHeader)
	_ = kong.Response.SetHeader("X-Country", countryHeader)
}

func main() {
	server.StartServer(New, "0.1", 1000)
}
