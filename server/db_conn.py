import psycopg2

ip = '127.0.0.1'

conn = psycopg2.connect("dbname=postgres user=admin")

# Open a cursor to perform database operations
cur = conn.cursor()

# Execute a command: this creates a new table
cur.execute("SELECT * FROM Users")
