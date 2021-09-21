#!/bin/sh

set -e
# Was the debug attribute set?
if [[ "${DEBUG}" == "true" ]]; then
    set -x
fi
echo "[$(date)]; DEBUG: ${DEBUG}"

echo "[$(date)]; Starting entrypoint.sh"

## Check required variables
DEFAULT_MINDMAX_VERSION=1.5.0
DEFAULT_GEOIP_UPDATE_VERSION=4.6.0
DEFAULT_GEOIP_MODULE_VERSION=3.3

if [ -z "${MAXMIND_LICENSE_KEY}" ]
then
    echo "[$(date)]; ERROR; The required MAXMIND_LICENSE_KEY environmental variable was not found, please refer to the README.md"
    exit 128
fi

if [ -z "${MAXMIND_ACCOUNT_ID}" ]
then
    echo "[$(date)]; ERROR; The required MAXMIND_ACCOUNT_ID environmental variable was not found, please refer to the README.md"
    exit 128
fi

echo "[$(date)]; Downloading and installing GeoLite2 Country DB tool"
curl -s -o /usr/share/geolite2_country_db.tar.gz "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-Country&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz"
tar -xf /usr/share/geolite2_country_db.tar.gz -C /usr/share/

echo "[$(date)]; Downloading and installing GeoLite2 City DB tool"
curl -s -o /usr/share/geolite2_city_db.tar.gz "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz"
tar -xf /usr/share/geolite2_city_db.tar.gz -C /usr/share/

mkdir -p /usr/local/share/GeoIP
mv /usr/share/GeoLite2*/GeoLite2*.mmdb /usr/local/share/GeoIP/

echo "[$(date)]; Removing downloaded artifacts"
rm -rf /usr/share/*.tar.gz
rm -rf /usr/share/GeoLite2-Country*
rm -rf /usr/share/GeoLite2-City*
rm -rf /usr/share/libmaxminddb-${MAXMIND_VERSION}

exec "$@"
