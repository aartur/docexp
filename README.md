# docexp - document expansion tool

`docexp` is a minimal documentation tool that expands code blocks in Markdown documents. It can be thought of as a simple replacement of Jupyter code cells, which doesn't use custom file formats or complex execution environments.

## How it works

Let's say you have a Markdown documentation file `docs.md` that includes a code block:

````markdown
### Database access

To inspect the database, run a sample query:

```sql
SELECT * FROM user LIMIT 3
```
````

Let's see how `docexp` parses the code block by using the `dry-run` command:

```bash
$ docexp dry-run docs.md
{"content":"SELECT * FROM user LIMIT 3\n","id":1,"metadata":{"column":0,"filename":"docs.md","headings":{"3":"Database access"},"lang":"sql","line":7}}
```

We can see the block was parsed as a JSON object, and we can access the SQL query, as well as check its file position and the Markdown headings hierarchy.

To run the query and produce some output, we need to create an *expander script*. The script will receive the code blocks shown by the `dry-run` command (as JSON-lines), and it should produce JSON objects `{"id": <id of the source block>, "exp": <expansion>}`.

A sample Python script `exp.py` looks like follows:

```python
#!/usr/bin/env python

import sys, json, subprocess

for line in sys.stdin:
    block = json.loads(line)
    match block.get("lang"):
        case "sql":
            res = subprocess.run(
                ["psql", "-d", "db_dev01", "-c", block["content"]],
                capture_output=True,
                text=True,
                check=True,
            )
            out = {"id": block["id"], "exp": res.stdout}
            print(json.dumps(out))
```

To run the expansion, execute:

```bash
$ docexp run --expander exp.py docs.md
```

`docs.md` now includes the expansion fragments:

````markdown
## Database access

You can also inspect the database content directly:

```sql
SELECT * FROM user LIMIT 3
```

<!-- <docexp> -->
```
 id  |  name  |       email        
-----+--------+--------------------
 123 | adam   | adam@example.com
 124 | eva    | eva@example.com
 125 | john   | john@example.com
(3 rows)
```
<!-- </docexp> -->

````

The expansion is marked using `<docexp>` tags within HTML comments, which ensures the regeneration of expansions will work correctly.

## Writing expander scripts

The tool moves much complexity to the expander script. You need to manage the various ways a code block can be executed:
- using a Docker container launched for the single command
- using a background Docker container to preserve the state and pass commands to it (that mimics Jupyter kernels)
- using different containers, depending on the `language` of a code block or a Markdown title hierarchy
- making security checks and using sandboxed environemnts (e.g. using a read-only database connection)

The future versions of the tool might include helpers for writing the expander scripts.

## Installation

Install the tool from the source code or from the releases page.

## Why someone would use `docexp`?

Full content/documentation tools like Jupyter Notebooks or Sphinx bring their own problems due to their size and complexity. They use custom file formats and heavy execution components (like Jupyter kernels).

`docexp` stays raw and simple. It's plain Markdown with one-line comment tags. The expander script concept makes code execution very explicit and traceable. The tool never modifies existing content, it only adds new content.

If all you need is code blocks expansion, `docexp` might work well for you.

## Implementation

The code doesn't include any AI-generated content. It's a Golang learning project.

The v0.1.* of the tool is a BETA release.

The implementation is fairly abstract with a domain model in `document.go`, 
which doesn't hard-code Markdown as the only supported format. It handles 
expandable spans being any file ranges, so it's possible to implement macro-like expansions of inline spans.
