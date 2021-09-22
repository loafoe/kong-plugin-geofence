# kong-plugin-geofence

Kong plugin implementing geofencing. Supports fencing at Country level currently. The plugin
uses the excellent [MaxMind](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) GeoIP2 or GeoLite2 databases.
The configured `license_key` will allow the plugin to download the database on-the-fly.

## configuration

```yaml
plugins:
- name: geofence
  config:
    license_key: XXX
    countries_allow_list:
    - NL
    - SR
```

## fields

* `config.license_key` - (Required) The MaxMind [license key](https://support.maxmind.com/account-faq/license-keys/how-do-i-generate-a-license-key/) to use
* `config.countries_allow_list` - (Optional) The list of country ISO codes to allow
* `config.countries_deny_list` - (Optional) The list of country ISO codes to deny

## license

License is MIT
