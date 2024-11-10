import psycopg2

ip = '127.0.0.1'
try:
    conn = psycopg2.connect("dbname=app user=admin password=password host="+ip+" port=6543")
    cursor = conn.cursor()
except ValueError:
    print('run the posgress and also setup script')

