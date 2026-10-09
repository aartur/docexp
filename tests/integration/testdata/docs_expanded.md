# Storing website content in an SQLite database

The documentation describes how to store website content into a local SQLite database, using Python.

## Make sure the website is valid

Use the `ping` command to make sure a website you selected is running:

```bash
ping -c1 example.com
```

<!-- <docexp> -->
```
PING example.com (104.20.23.154) 56(84) bytes of data.
64 bytes from 104.20.23.154: icmp_seq=1 ttl=51 time=38.8 ms

--- example.com ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 38.839/38.839/38.839/0.000 ms
```
<!-- </docexp> -->

## Create and fill the database

Create the SQLite database using the CLI tool:

```bash
sqlite3 website.db "CREATE TABLE IF NOT EXISTS website (url TEXT, status_code INT, content TEXT)"
```

<!-- {"tool": "docexp", "tag": "expansion_start", "expander": "exp.sh", "started": "2026-08-11T10:23:55+02:00", elapsed_seconds: 10.0} -->

<!-- {"tool": "docexp", "tag": "expansion_end"} -->

Now we can load sample content using Python:

```python
import sqlite3, requests

r = requests.get('http://example.com')
conn = sqlite3.connect('website.db')
cur = conn.cursor()
cur.execute('INSERT INTO website VALUES (?, ?, ?)', ('example.com', r.status_code, r.text))
print(f"Inserted {cur.rowcount} rows")
conn.commit()
conn.close()
```

<!-- {"tool": "docexp", "tag": "expansion_start", "expander": "exp.sh", "started": "2026-08-11T10:23:55+02:00", elapsed_seconds: 10.0} -->
```
Inserted 1 rows
```
<!-- {"tool": "docexp", "tag": "expansion_end"} -->

## Query the Data

Query the database to make sure the expected content is present:

```sql
SELECT COUNT(*) FROM website WHERE content like '%Example Domain%'
```
<!-- {"tool": "docexp", "tag": "expansion_start", "expander": "exp.sh", "started": "2026-08-11T10:23:55+02:00", elapsed_seconds: 10.0} -->
```
1
```
<!-- {"tool": "docexp", "tag": "expansion_end"} -->
