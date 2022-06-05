# memcached
Memcached driver

### Step 1

make build

### Step 2

make run

### STOP

make stop

## Example of the configuration used

Below is a list of variables that should be used to run the project:

| ENV VAR             | Example Value                      | Description                                 | Default value   |
|---------------------|------------------------------------|---------------------------------------------|-----------------|
| MCD_IP              | `'0.0.0.0'`                        | server host                                 | `0.0.0.0`       |
| MCD_PORT            | `'8801'`                           | server port                                 | `8801`          |
| MCD_REPOSITORY_TYPE | `'1'`                              | repository type: 1 - memcached, 2- inmemory | `1`             |
| MCD_LOGLEVEL        | `'warn'`                           | log level                                   | `warn`          |
| MCD_MEMCACHED_ADR   | `"dGpoDoOVrNQxNHPOSMeijkbYmAJTSF"` | memcached address                           | `0.0.0.0:11211` |
| MCD_MAX_OPEN_CONN   | `"500"`                            | max open conn                               | `500`           |
| MCD_MAX_IDLE_CONN   | `"500"`                            | max idle conn                               | `500`           |