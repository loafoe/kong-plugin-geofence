package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"sync"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
	"github.com/loafoe/mmdb"
	"github.com/oschwald/geoip2-golang"
)

// Config
type Config struct {
	LicenseKey         string   `json:"license_key"`
	CountriesAllowList []string `json:"countries_allow_list"`
	CountriesDenyList  []string `json:"countries_deny_list"`
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

func contains(s []string, searchTerm string) bool {
	i := sort.SearchStrings(s, searchTerm)
	return i < len(s) && s[i] == searchTerm
}

// Access implements the Access step
func (conf Config) Access(kong *pdk.PDK) {

	doOnce.Do(func() {
		db, dbErr = initDB(conf.LicenseKey)
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
	countryCode := fmt.Sprintf("%v", record.Country.IsoCode)
	_ = kong.ServiceRequest.SetHeader("X-Geofence-Detected-Country", countryCode)

	block := false
	// Filter checks
	if len(conf.CountriesAllowList) > 0 && !contains(conf.CountriesAllowList, countryCode) {
		block = true
	}
	if len(conf.CountriesDenyList) > 0 && contains(conf.CountriesDenyList, countryCode) {
		block = true
	}
	if block {
		headers := map[string][]string{}
		kong.Response.Exit(http.StatusUnauthorized, "blocked\n", headers)
		return
	}
}

func main() {
	_ = server.StartServer(New, "0.1", 1000)
}
