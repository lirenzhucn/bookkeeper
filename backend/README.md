# Backend and CLI for Personal Finance Bookkeeping

In general, the user uses `bkpctl` to interact with the system. And it requires
the API server `bkpsrv` to be up and running. To start the server, you need to
- Have a PostgreSQL database service running
- Specify database URL (including username and password) in `./configs/bkpsrv.yml`
You can run the following command to start the server:
```
go run ./cmd/bkpsrv
```

## Database Manipulation
These operations require direct access to the PostgreSQL database and,
therefore, will only be feasible on the backend server.

```
go run ./cmd/bkpctl db init
```

**DANGER**: This command **wipes** and initializes the database.

## Import Data
Currently the system supports the imoprt of the data that are exported by the
sui.com iOS app (随手记专业版) and in csv format. To import the data, you also
need to specify some additional info in a JSON config file. The config file
contains specification of all accounts, the date/time format string, and the
translation maps of categories and transaction types, as sui.com data are in
Chinese.

To imoprt the data, run:
```
go run ./cmd/bkpctl import -c </path/to/config.json> -d <path/to/data.csv>
```

## Financial Statements
The system supports generation of Balance Sheets and Income Statements for
multiple dates and periods. Some feature highlights are:
- Display multiple dates and periods side by side in a table
- Customize the tags and categories to collect in the statements
- Customize how the data is presented by specifying a report schema
- Use arbitrary dates and periods, as well as shorthands like 2021Q1 and 2022H1