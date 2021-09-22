package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/loafoe/mmdb"
	"github.com/oschwald/geoip2-golang"
)

// Config
type Config struct {
	LicenseKey string `json:"license_key"`
	AllowList  string `json:"allow_list"`
	DenyList   string `json:"deny_list"`
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
var denyList []string
var allowList []string

var doOnce sync.Once
var dbErr error

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

// Access implements the Access step
func (conf Config) Access(kong *pdk.PDK) {

	doOnce.Do(func() {
		db, dbErr = initDB(conf.LicenseKey)
		if len(conf.AllowList) > 0 {
			allowList = strings.Split(conf.AllowList, ",")
		}
		if len(conf.DenyList) > 0 {
			denyList = strings.Split(conf.DenyList, ",")
		}
	})

	if db == nil {
		_ = kong.ServiceRequest.SetHeader("X-Geofence-Detected-Country-Error", fmt.Sprintf("GeoIP database not ready: %v", dbErr))
		headers := map[string][]string{
			"X-Kong-Geofence": {"active"},
		}
		kong.Response.Exit(http.StatusTooManyRequests, fmt.Sprintf("GeoIP database not ready: %v", dbErr), headers)
		return
	}
	clientIP, err := kong.Client.GetForwardedIp()
	if err != nil {
		_ = kong.ServiceRequest.SetHeader("X-Geofence-Detected-Country-Error", err.Error())
		return
	}
	_ = kong.ServiceRequest.SetHeader("X-Geofence-Detected-IP", clientIP)
	ip := net.ParseIP(clientIP)
	record, err := db.Country(ip)
	if err != nil {
		_ = kong.ServiceRequest.SetHeader("X-Geofence-Detected-Country-Error", err.Error())
		return
	}
	countryHeader := fmt.Sprintf("%v", record.Country.IsoCode)
	_ = kong.ServiceRequest.SetHeader("X-Geofence-Detected-Country", countryHeader)

	// Filter checks
	if len(allowList) > 0 && contains(allowList, countryHeader) {
		return
	}
	if len(denyList) > 0 && contains(denyList, countryHeader) {
		headers := map[string][]string{}
		kong.Response.Exit(http.StatusUnauthorized, fmt.Sprintf("%s is blocked", countryHeader), headers)
		return
	}
}

func main() {
	_ = server.StartServer(New, "0.1", 1000)
}
