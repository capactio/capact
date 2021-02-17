# Jinja-cli

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

* support for passing multiple files with data,
* functions to generate random string:
  - `def random_string(letters: str = "", length: int = 10) -> str`
    using passed `letters`
    If no letters were provided it is using all printable letters
    except whitespaces and quotes.
  - `random_word(length: int = 10) -> str`
    generates random word of the given length using only
    lowercase asci letters.
* missing variables are not causing errors anymore. Template can be rendered
  several times,
* for variables use now `< variable >` instead of `{{ variable }}`,
* for blocks use now `<% block %>` instead of `{% block %}`.

Jinja cli is a fork of https://github.com/mattrobenolt/jinja2-cli
Docker part is a fork from https://github.com/dinuta/jinja2docker

## Prerequisites

- [Python](https://python.org)

## Setup

Setup Python environment.

```bash
python3 -m venv /tmp/jinja
source /tmp/jinja/bin/activate
pip install wheel
pip install -e jinja2-cli[yaml]
```

## Usage

Run:

```bash
jinja2 testdata/user.tmpl testdata/data1.yaml testdata/data2.yaml
```
