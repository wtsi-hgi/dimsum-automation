# dimsum-automation
Automated running of dimsum

## Deployment
To connect to the google sheet with experimental data, you'll need the service
credentials JSON file, the path to which can be specified in an environment
variable, along with the spreadsheet ID and mlwh database connection details
(read-only): 

```
export DIMSUM_AUTOMATION_CREDENTIALS_FILE=/path/to/credentials.json
export DIMSUM_AUTOMATION_SPREADSHEET_ID="1ksldfj3lsadfj"
export DIMSUM_AUTOMATION_SQL_HOST=localhost
export DIMSUM_AUTOMATION_SQL_PORT=3306
export DIMSUM_AUTOMATION_SQL_USER=user
export DIMSUM_AUTOMATION_SQL_PASS=pass
export DIMSUM_AUTOMATION_SQL_DB=mlwarehouse
```

If you put these statements in a `.env` file that's in the current working
directory when you start the server, it will automatically be sourced.

