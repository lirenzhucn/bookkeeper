# Bookkeeper controler utility

The bookkeeper control (`bkpctl`) provides a CLI to interact with the bookkeeper
server.

```
bkpctl transactions list --start 20210101 --end 20210331
```
This command lists all transactions between Jan 1, 2021 and March 31, 2021

```
bkpctl import-sui --data </path/to/sui.com/export.csv> --config </path/to/config.json>
```
This command tries to import the data file exported by sui.com (随手记) iOS app,
with the configuration. (TODO: define how the configuration should be
specified in the JSON file.)
