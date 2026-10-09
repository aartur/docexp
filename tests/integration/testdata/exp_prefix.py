#!/usr/bin/env python

import sys, json

for line in sys.stdin:
    span = json.loads(line)
    exp = {"id": span["id"], "exp": f'exp: {span["content"]}'}
    print(json.dumps(exp))
