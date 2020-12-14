**Grafana to Matrix Forwarder**

Forward alerts from [Grafana](https://grafana.com) to a [Matrix](https://matrix.org) chat room.

 [![pipeline status](https://gitlab.com/hectorjsmith/grafana-matrix-forwarder/badges/main/pipeline.svg)](https://gitlab.com/hectorjsmith/grafana-matrix-forwarder/-/commits/main) [![Go Report Card](https://goreportcard.com/badge/gitlab.com/hectorjsmith/grafana-matrix-forwarder)](https://goreportcard.com/report/gitlab.com/hectorjsmith/grafana-matrix-forwarder)

---

This project provides a simple way to forward alerts generated by Grafana to a Matrix chat room.

Setup up a Grafana webhook alert channel that targets an instance of this application.
This tool will handle converting the incoming alert webhook to a Matrix message and send it on to a specific chat room.

![screenshot of matrix alert message](docs/alertExample.png)

## Features

  * 📦 **Portable**
    * As a single binary the tool is easy to run in any environment
  * 📎 **Simple**
    * No config files, all required parameters provided on startup
  * 🪁 **Flexible**
    * Support multiple grafana alert channels to multiple matrix rooms
  * 📈 **Monitorable**
    * Export metrics to track successful and failed forwards

## How to use

**Step 1**

Run the forwarder by providing a matrix account to send messages from.

```
$ ./grafana-matrix-forwarder --user @userId:matrix.org --password xxx --homeserver matrix.org
```

**Step 2**

Add a new **POST webhook** alert channel with the following target URL: `http://<ip address>:6000/api/v0/forward?roomId=<roomId>`

*Replace with the server ID and matrix room ID.*

![screenshot of grafana channel setup](docs/grafanaChannelSetup.png)

**Step 3**

Setup alerts in grafana that are sent to the new alert channel.

![screenshot of grafana alert setup](docs/grafanaAlertSetup.png)

## CLI Usage

```
$ grafana-matrix-forwarder -h

  -homeserver string
        url of the homeserver to connect to (default "matrix.org")
  -host string
        host address the server connects to (default "0.0.0.0")
  -logPayload
        print the contents of every alert request received from grafana
  -password string
        password used to login to matrix
  -port int
        port to run the webserver on (default 6000)
  -resolveMode string
        set how to handle resolved alerts - valid options are: 'message', 'reaction' (default "message")
  -user string
        username used to login to matrix
  -version
        show version info and exit
``` 

## Metrics

Access exported metrics at `/metrics` (on the same port). Metrics are compatible with prometheus.

**Note:** All metric names include the `gmf_` prefix (grafana matrix forwarder) to make sure they are unique and make them easier to find.

Exposed metrics:
  * `up` - Returns 1 if the service is up
  * Forward counts
    * `total` - total number of alerts forwarded
    * `success` - number of alerts successfully forwarded
    * `error` - number of alerts where the forwarding process failed (check logs for error details)
  * Alert counts by state
    * `total` - total number of alerts received
    * `alerting` - alert count in the *alerting* state
    * `no_data` - alert count in the *no_data* state
    * `ok` - alert count in the *ok* state (resolved alerts)
    * `other` - number of received alerts that have an unknown state (check logs for details)

**Sample**

```
# HELP gmf_up
# TYPE gmf_up gauge
gmf_up 1
# HELP gmf_forwards
# TYPE gmf_forwards gauge
gmf_forwards{result="error"} 1
gmf_forwards{result="success"} 5
gmf_forwards{result="total"} 6
# HELP gmf_alerts
# TYPE gmf_alerts gauge
gmf_alerts{state="alerting"} 1
gmf_alerts{state="no_data"} 1
gmf_alerts{state="ok"} 2
gmf_alerts{state="other"} 1
gmf_alerts{state="total"} 6
```

## Thanks

Made possible by the [maunium.net/go/mautrix](https://maunium.net/go/mautrix/) library and all the contributors to the [matrix.org](https://matrix.org) protocol.
