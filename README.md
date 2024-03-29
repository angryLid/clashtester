<h1 align="center">Clash Tester</h1>
<h4 align="center">A clash speed test program.</h4>

## Feature
* Test latency, download and upload speed via speed.cloudflare.com
* generate excel report

## Usage

1. Switch to subscription that you are going to test.
2.  download from [Releases](https://github.com/angryLid/clashtester/releases) and run following command:

```bash 
go run main.go [name]
# eg. go run main.go us jp
# It will test all proxies whose name includes us or jp.
```
## How It Works
1. Download 100KB x 8
2. Upload 100KB x 8
3. Download 10MB x 2

## Known Issues

if could't download 10MB data within 20 seconds, then it will report 0.00MB/s. But the speed is probably between 0 and 0.5 MB/s

## References

[KNawm/speed-cloudflare-cli](https://github.com/KNawm/speed-cloudflare-cli/)