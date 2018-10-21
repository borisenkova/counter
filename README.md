only linux file paths are supported
stdin separator is '\n' (*nix newline)
only successfully processed inputs are included in 'total' in the end
max number of 'Go' in Source are limited by uint64.
max number of workers are limited by int (please be sure that ulimit -n returns reasonable number before increasing number of workers in config).
min number for requesting and reading content from URL is 1 second, max -- int64 seconds. The timeout could be specified only in seconds.
supported schemes for urls are HTTP and HTTPS only.