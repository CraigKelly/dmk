#!/usr/bin/env python3

"""Some helper functions for python scripts in this project."""

import inspect
import os
import subprocess
import sys

pth = os.path


FILE_PATH = inspect.getabsfile(lambda i: i)
SCRIPT_DIR = pth.dirname(FILE_PATH)


def rel_path(rel):
    """Return a path relative to this script's dir."""
    return pth.abspath(pth.join(SCRIPT_DIR, rel))


def flush():
    """Flush and sync to keep everything straight."""
    sys.stdout.flush()
    sys.stderr.flush()
    os.sync()


def cmd(cl, *args):
    """Run command-line, possibly with string interpolation."""
    if args:
        cl = cl % args
    print(cl)
    flush()
    return subprocess.call(cl.split(' '))
    flush()
