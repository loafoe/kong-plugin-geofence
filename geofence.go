package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/Kong/go-pdk"
	"github.com/loafoe/mmdb"
	"github.com/oschwald/geoip2-golang"
)

// Config
type Config struct {
}

//nolint
func New() interface{} {
	return &Config{}
}

var dbMutex sync.Mutex

func InitDB() (*geoip2.Reader, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	file, err := ioutil.TempFile("", "*.mmdb")
	if err != nil {
		dbErr = err
		return nil, err
	}
	err = mmdb.Download(file.Name(), os.Getenv("LICENSE_KEY"))
	if err != nil {
		dbErr = err
		return nil, err
	}
	reader, err := geoip2.Open(file.Name())
	if err != nil {
		dbErr = err
		return nil, err
	}
	return reader, nil
}

var db *geoip2.Reader
var doOnce sync.Once
var dbErr error

// Access implements the Access step
func (conf *Config) Access(kong *pdk.PDK) {
	doOnce.Do(func() {
		db, dbErr = InitDB()
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
