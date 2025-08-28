# kong-plugin-geofence

Kong plugin implementing geofencing. Supports fencing at Country level currently. The plugin
uses the excellent [MaxMind](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) GeoIP2 or GeoLite2 databases.
The configured `license_key` will allow the plugin to download the database on-the-fly.

## Kong environment

```shell
KONG_PLUGINS = "geofence,bundled"
KONG_PLUGINSERVER_NAMES = "geofence"
KONG_PLUGINSERVER_GEOFENCE_START_CMD = "/usr/local/bin/geofence"
KONG_PLUGINSERVER_GEOFENCE_QUERY_CMD = "/usr/local/bin/geofence -dump"
```

## Kong configuration

### Declarative

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

* `config.license_key` - (Required) The MaxMind [license key](https://support.maxmind.com/hc/en-us/articles/4407111582235-Generate-a-License-Key) to use
* `config.countries_allow_list` - (Optional) The list of country ISO codes to allow
* `config.countries_deny_list` - (Optional) The list of country ISO codes to deny

## license

License is MIT
