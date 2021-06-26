# MyBack MySQL backup client & server

MyBack is a multi-platform client/server utility to back up a MySQL/MariaDB database over HTTP(S). 


## Features

- Server uses MySQL/MariaDB authentication for client access, can be limited to certain user(s) and/or IP(s)
- Optional server TLS encryption via HTTPS
- Server uses native `mysqldump` over TCP to stream data to client
- Client only downloads changed or modified tables over HTTP with gzip compression
- Client option to store backups compressed in [zstd format](https://facebook.github.io/zstd/) (default true)
- Client option to specify which databases to back up (supports wildcards)
- Client option to ignore databases, tables or table-data (supports wildcards)
- Client option to selectively dump only a subset of specified table data based on SQL query
- Individual table dumps stored on client side, can be merged into single SQL file (see `myback extract -h`)


## Limitations

- Given the nature of the selective backup process, databases are not locked during dumps
- The MySQL/MariaDB server must be accessible to the server via TCP


## Server requirements

- A running server (see `myback server -h` for options) on a port of your choice (default 3307). 
- The machine running the server must also have `mysqldump` in the $PATH, and be able to access the database over TCP.


## Client requirements

- A valid yaml configuration file
- A username/password of a user authorized to dump tables from the server


## Server configuration 

The server is started with command-line arguments and will run in the foreground:

```
myback server -h

Usage:
  myback server [flags]

Flags:
  -h, --help                    help for server
      --ips string              limit to ips (comma-separated)
      --listen string           listen on interface:port (default "0.0.0.0:3307")
      --mysql-host string       MySQL server host (default "localhost")
      --mysql-port int          MySQL server port (default 3306)
      --mysqldump string        mysqldump command (default "mysqldump")
      -t, --show-timestamps     show timestamps in log output
      --ssl-cert string         SSL certificate (optional, must be used with --ssl-key)
      --ssl-key string          SSL private key (optional, must be used with --ssl-cert)
      --users string            limit to users (comma-separated)
  -v, --verbose                 verbose output
```

Examples:

```
# start the server with default values, database local, no https
myback server

# start the server using HTTPS - note you will need to restart the MyBack server if you update the certificate
myback server --ssl-cert /etc/letsencrypt/live/example.com/fullchain.pem --ssl-key /etc/letsencrypt/live/example.com/privkey.pem

# limit the server to root & dumpuser
myback server --users root,dumpuser
```

MyBack server can be automatically started via systemd (see the example [`contrib/myback.service`](contrib/myback.service)).


## Client configuration

The client requires a valid yaml file (see the example [`contrib/client-example.yml`](contrib/client-example.yml)).

```
# myback -h

Usage:
  myback backup <client-config> [flags]

Flags:
  -h, --help      help for backup
  -v, --verbose   verbose output
```


## Extracting backups

Backups are stored in individual files (either compressed or not depending on client configuration). 
MyBack can generate complete SQL files using the `extract` option, and will automatically include any backup files found beneath the provided path(s).

```
# generate a full SQL dump of entire server and save to /tmp/full.sql
myback extract /mnt/backups/mysql -o /tmp/full.sql

# generate a SQL file from selected databases and save to /tmp/selected.sql
myback extract /mnt/backups/mysql/mydatabase /mnt/backups/mysql/myotherdatabase -o /tmp/selected.sql
```


## Backing up database dumps

MyBack does not keep multiple versions of each backup, so you are responsible for your own backup storage & rotation. 

Personally I use [restic](https://restic.net) which works perfectly for me.
