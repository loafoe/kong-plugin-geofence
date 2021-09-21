package main

import (
	"fmt"
	"io/ioutil"
	"net"
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

func init() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	file, err := ioutil.TempFile("", "*.mmdb")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	err = mmdb.Download(file.Name(), os.Getenv("LICENSE_KEY"))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	db, err = geoip2.Open(file.Name())
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Printf("Done: %s\n", file.Name())
}

var db *geoip2.Reader

// Access implements the Access step
func (conf Config) Access(kong *pdk.PDK) {
	if db == nil {
		_ = kong.ServiceRequest.SetHeader("X-Detected-Country-Error", "GeoIP database not ready")
		//headers := map[string][]string{
		//	"X-Kong-Geofence": {"active"},
		//}
		//kong.Response.Exit(http.StatusForbidden, "No GeoIP database", headers)
		return
	}
	clientIP, err := kong.Client.GetIp()
	if err != nil {
		_ = kong.ServiceRequest.SetHeader("X-Detected-Country-Error", err.Error())
		_ = kong.Log.Err(err.Error())
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
