# faas-grafana-annotate
An OpenFaaS function to create annotations in Grafana (>= v4.6).

Install faas-cli
```bash
curl -sSL https://cli.openfaas.com | sudo sh
```

## Supported Request Parameters
- body (string, text/plain)
- query string
  - tag (string, multiple, optional)
  - panelId (int, optional)
  - dashboardId (int, optional)

By default, if no tags are provided, the tag "global" is used for your annotation. In your dashboard, make sure *Annotations & Alerts* is enabled and filtered by the appropriate tag.

## Function Configuration
- environment
  - grafana_url
- secrets
  - grafana_api_token
  - grafana_username
  - grafana_password

This function prioritizes **grafana_api_token** first, then falls back to basic authentication provided by grafana_username and grafana_password.

### Grafana Configuration
1) Add new API key (User icon -> API Keys) w/ Editor role
2) In your dashboard, click gear icon -> Annotations -> enable
   > In your dashboard, make sure *Annotations & Alerts* is enabled and and filtered by the appropriate tag.

## Deployment Examples
faas-cli (from Docker)
```bash
faas-cli deploy --image acornies/grafana-annotate --env grafana_url=http://example:3000
```
faas-cli (from source)
```bash
faas-cli deploy -f ./grafana-annotate.yml --env grafana_url=http://example:3000
```

## Invoke Examples
faas-cli
```bash
faas-cli invoke grafana-annotate --query tag=global --query tag=faas --query dashboardId=1 --query panelId=1 --gateway http://localhost:8080
```
curl
```bash
curl -XPOST -d 'test annotation' "http://localhost:8080/function/grafana-annotate?tag=global&tag=faas&tag=application"
```
## Screenshot Example
![faas grafana screenshot](https://github.com/tucows/faas-grafana-annotate/blob/master/grafana_screen.png) 
