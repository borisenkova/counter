# Counter

Counts substring "Go" in sources from stdin.

## Limitations

* only Go standard library is used for implementation. `github.com/stretchr/testify` is used in tests;
* no global variables are used;
* only linux file paths are supported;
* stdin separator is '\n' (*nix newline);
* only successfully processed sources are included in 'Total' in the end;
* max number of workers are limited by math.MaxInt32 (please be sure that `ulimit -n` returns reasonable number before increasing the number of workers);
* min duration for requesting and reading content from URL is 1 millisecond, max -- math.MaxInt64 milliseconds. The timeout could be specified only in milliseconds;
* supported schemes for URLs are **HTTP** and **HTTPS** only. URL without scheme is considered invalid.

## Config parameters

Parameters are passed through environmental variables. Possible parameters are:
* `MAX_NUMBER_OF_WORKERS` -- maximum number of sources to process simultaneously. Should be above zero and below math.MaxInt32. Default value is `5`;
* `SUBSTRING` -- string to count in sources. Default value is `Go`;
* `URL_REQUEST_TIMEOUT_MILLISECONDS` -- timeout for processing URL source in milliseconds. Default value is `1000`. Should be increased before processing large web pages or in case of slow network or slow endpoint servers;

## Example

```
$ echo -e 'https://golang.org\n/etc/passwd\nhttps://golang.org\nhttps://golang.org' | MAX_NUMBER_OF_WORKERS=1000 go run 1.go
Count for /etc/passwd: 0
Count for https://golang.org: 8
Count for https://golang.org: 8
Count for https://golang.org: 8
Total: 24
```
