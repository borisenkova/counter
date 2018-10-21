only linux file paths are supported
stdin separator is '\n' (*nix newline)
only successfully processed inputs are included in 'total' in the end
max number of 'Go' in Source are limited by uint64.
max number of workers are limited by int (please be sure that ulimit -n returns reasonable number before increasing number of workers in config).
min duration for requesting and reading content from URL is 1 millisecond, max -- int64 milliseconds. The timeout could be specified only in milliseconds.
supported schemes for urls are HTTP and HTTPS only.