#!/usr/bin/env python

import sys, json

for line in sys.stdin:
    span = json.loads(line)
    print(json.dumps({"id": span["id"], "exp": span["content"]}))