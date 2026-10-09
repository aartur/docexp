# Storing website content in an SQLite database

The documentation describes how to store website content into a local SQLite database, using Python.

## Make sure the website is valid

Use the `ping` command to make sure a website you selected is running:

```bash
ping -c1 example.com
```

## Create and fill the database

Create the SQLite database using the CLI tool:

```bash
sqlite3 website.db "CREATE TABLE IF NOT EXISTS website (url TEXT, status_code INT, content TEXT)"
```

Now we can load sample content using Python:

```python
import sqlite3, requests

r = requests.get('http://example.com')
conn = sqlite3.connect('website.db')
conn.cursor().execute('INSERT INTO website VALUES (?, ?, ?)', ('example.com', r.status_code, r.text))
conn.commit()
conn.close()
```

## Query the Data

Query the database to make sure the expected content is present:

```sql
SELECT COUNT(*) FROM website WHERE content like '%Example Domain%'
```

