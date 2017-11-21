# OSCtoInfluxDB
OSC server that writes messages to influx server running on the same machine.

# influxdb

docker run --rm influxdb influxd config > influxdb.conf

//Edit udp section in file

docker run -d --name=influx -p 8086:8086 -p 8089:8089/udp -v $PWD\influxdb.conf:/etc/influxdb/influxdb.conf:ro influxdb -config /etc/influxdb/influxdb.conf


# grafana

docker run -d --name=grafana -p 3000:3000 grafana/grafana --link influxdb

