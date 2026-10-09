# docexp - document expansion tool

A minimal documentation tool that expands code blocks in Markdown documents.

## How it works

Let's say you have a Markdown documentation file `docs.md` that includes two code blocks:

````markdown
## Users

To get the registered users, execute:

```python
from app import models

users = models.User.objects.all()
print(users[0])
```

### Database access

You can also inspect the database content directly:

```sql
SELECT * FROM user LIMIT 3
```
````

And now you would like to execute the code inside the blocks and include its output in the documentation file. Let's use the `docexp dry-run` command which parses the file and shows input for an *expansion script*:

```bash
$ ./docexp dry-run tests/integration/testdata/doc_readme.md
{"content":"from app import models\n\nusers = models.User.objects.all()\nprint(users[0])\n","id":1,"metadata":{"column":0,"filename":"tests/integration/testdata/doc_readme.md","headings":{"2":"Users"},"lang":"python","line":4}}
{"content":"SELECT * FROM user LIMIT 3\n","id":2,"metadata":{"column":0,"filename":"tests/integration/testdata/doc_readme.md","headings":{"2":"Users","3":"Database access"},"lang":"sql","line":15}}
```

The format is JSON lines. Each line describes a code block by including its content, file position and Markdown headings hierarchy. The expansion script `exp.py` consumes such input and expands each block by executing its content:

```python
#!/usr/bin/env python

import sys, json, subprocess, io, contextlib

for line in sys.stdin:
    block = json.loads(line)
    match block.get("lang"):
        case "python":
            buffer = io.StringIO()
            with contextlib.redirect_stdout(buffer):
                exec(code_fragment)
            out = {"id": block["id"], "exp": buffer.getvalue()}
            print(json.dumps(out))
        case "sql":
            res = subprocess.run(
                ["psql", "-d", "db01", "-c", block["content"]],
                capture_output=True,
                text=True,
                check=True,
            )
            out = {"id": block["id"], "exp": res.stdout}
            print(json.dumps(out))
```

To run the expansion, execute:

```bash
docexp run --expander exp.py docs.md
```

`docs.md` now includes the expansion fragments:

````markdown
## Users

To get the registered users, execute:

```python
from app import models

users = models.User.objects.all()
print(users[0])
```

<!-- <docexp> -->
```
<User id=123 name=adam email=adam@example.com>
```
<!-- </docexp> -->

### Database access

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

## Writing expander scripts

The expander script can be a complex piece of code. Usually it needs to preserve state during execution, so you might need to do things like running Docker containers in background and passing code blocks to execute. You must also handle security and detecting errors.

## Installation

Install the tool from the source code or from the releases page.

## Why someone would use `docexp`?

Full content/documentation tools like Jupyter Notebooks or Sphinx bring their own complexity due to their size. They use custom file formats and use heavy execution components (like Jupyter kernels).

`docexp` stays raw and simple. Both inputs and outputs are plain Markdown documents. The expander script concept makes code execution very explicit and traceable. The tool never modifies existing content, it can only add new content.

If all you need is code blocks expansion, `docexp` might work well for you.

## Implementation

The code doesn't include any AI-generated content. It's a Golang learning project.

The implementation is fairly abstract with a domain model in `document.go`, which doesn't
hard-code Markdown as the only supported format. It handles expandable spans being any file ranges.