#!/usr/bin/python3

import urllib2
import re
import sys

level = sys.argv[1]
set_id = sys.argv[2]
url = "http://view.websudoku.com/?level=%s&set_id=%s" % (level, set_id)

f = urllib2.urlopen(url).read()
mask = re.findall('editmask.*VALUE="(.*)"', f)[0]
cheat = re.findall('cheat.*VALUE="(.*)"', f)[0]

puzzle = ""
solution = ""
for i, c, m in zip(range(81), cheat, mask):
    puzzle += (c if m == "0" else "_")
    solution += c
    sep = " " if (i+1)%9 != 0 else "\n"
    puzzle += sep
    solution += sep

print("// %s\n" % url)
print("Level: %s\n" % level)
print("Puzzle:\n%s" % puzzle)
print("Solution:\n%s" % solution)
