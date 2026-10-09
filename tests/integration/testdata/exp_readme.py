#!/usr/bin/env python

import sys, json, subprocess, io, contextlib

for line in sys.stdin:
    block = json.loads(line)
    match block.get("lang"):
        case "python":
            buffer = io.StringIO()
            with contextlib.redirect_stdout(buffer):
                exec(block["content"])
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
