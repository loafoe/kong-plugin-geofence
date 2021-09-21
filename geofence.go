package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/Kong/go-pdk"
	"github.com/oschwald/geoip2-golang"
)

const (
	downloadUrl = "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=%s&suffix=tar.gz"
)

// Config
type Config struct {
	LicenseKey string `json:"licenseKey"`
}

// New
func New() interface{} {
	return &Config{}
}

var _ = New()

var doOnce sync.Once
var db *geoip2.Reader

// Access implements the Access step
func (conf Config) Access(kong *pdk.PDK) {
	doOnce.Do(func() {
		file, err := ioutil.TempFile("", "*.mmdb")
		if err != nil {
			_ = kong.Log.Err(err.Error())
			return
		}

		downloadURL := fmt.Sprintf(downloadUrl, conf.LicenseKey)
		_ = kong.Log.Debug("Downloading GeoIP database...")
		resp, err := http.Get(downloadURL)
		if err != nil {
			_ = kong.Log.Err(err.Error())
			return
		}
		defer resp.Body.Close()
		uncompressedStream, err := gzip.NewReader(resp.Body)
		if err != nil {
			_ = kong.Log.Err(err.Error())
			return
		}
		tarReader := tar.NewReader(uncompressedStream)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				_ = kong.Log.Err(err.Error())
				return
			}
			switch header.Typeflag {
			case tar.TypeDir:
				continue
			case tar.TypeReg:
				if filepath.Ext(header.Name) != ".mmdb" { // Skip all but the database
					continue
				}
				outFile, err := os.Create(file.Name())
				if err != nil {
					_ = kong.Log.Err(err.Error())
					return
				}
				if _, err := io.Copy(outFile, tarReader); err != nil {
					_ = kong.Log.Err(err.Error())
					return
				}
				outFile.Close()
			default:
				err = fmt.Errorf(
					"ExtractTarGz: uknown type: %v in %s",
					header.Typeflag,
					header.Name)
				_ = kong.Log.Err(err.Error())
				return
			}
		}

		db, err = geoip2.Open(file.Name())
		if err != nil {
			_ = kong.Log.Err(err.Error())
		}
	})
	if db == nil {
		_ = kong.Response.SetHeader("X-Detected-Country-Error", "No GeoIP database")
		//headers := map[string][]string{
		//	"X-Kong-Geofence": {"active"},
		//}
		//kong.Response.Exit(http.StatusForbidden, "No GeoIP database", headers)
		return
	}
	clientIP, err := kong.Client.GetIp()
	if err != nil {
		_ = kong.Response.SetHeader("X-Detected-Country-Error", err.Error())
		_ = kong.Log.Err(err.Error())
		return
	}
	_ = kong.Response.SetHeader("X-Detected-IP", clientIP)
	ip := net.ParseIP(clientIP)
	record, err := db.Country(ip)
	if err != nil {
		_ = kong.Response.SetHeader("X-Detected-Country-Error", err.Error())
		return
	}
	countryHeader := fmt.Sprintf("[%v]", record.Country.IsoCode)
	_ = kong.ServiceRequest.SetHeader("X-Detected-Country", countryHeader)
	_ = kong.Response.SetHeader("X-Country", countryHeader)
}
