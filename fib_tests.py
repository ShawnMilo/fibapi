#!/usr/bin/env python3

import sys
import json
from json.decoder import JSONDecodeError

import requests

url = 'http://localhost:8000/'

ord_tests = (
    (1, 1),
    (2, 1),
    (3, 2),
    (4, 3),
    (5, 5),
    (6, 8),
    (12, 144),
    (20, 6765),
)

below_tests = (
    (1, 0),
    (2, 2),
    (3, 3),
    (4, 4),
    (6, 5),
    (7, 5),
    (8, 5),
    (9, 6),
)

def get(url):
    resp = requests.get(url)
    try:
        return json.loads(resp.content)
    except JSONDecodeError:
        sys.stderr.write(f'failed to decode JSON from {resp.content} \n')
    return {}

def ord_call():
    for num, expected in ord_tests:
        data = get(url+ f'ordinal/{num}')
        found = data.get('num')
        if found != expected:
            print(data)
            sys.stderr.write(f'ord {num} expected {expected}, got {found}\n')
            return False
    return True

def below_call():
    for num, expected in below_tests:
        data = get(url+ f'below/{num}')
        found = data.get('count')
        if found != expected:
            print(data)
            sys.stderr.write(f'below {num} expected {expected}, got {found}\n')
            return False
    return True

def below20():
    data = get(url+ 'below/20')
    if data.get('count') != 7:
        return False
    return True

def below1():
    data = get(url+ 'below/1')
    if data.get('count') != 0:
        return False
    return True

def main():
    ord_call()
    below_call()

if __name__ == '__main__':
    main()


