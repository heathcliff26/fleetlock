package sql

func createConnectionString(username, password, address, database, options string) string {
	var connStr string
	if username != "" {
		connStr += username
		if password != "" {
			connStr += ":" + password
		}
		connStr += "@"
	}
	connStr += address + "/" + database
	if options != "" {
		connStr += "?" + options
	}

	return connStr
}
