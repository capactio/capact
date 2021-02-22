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
* for variables use now `<@ variable @>` instead of `{{ variable }}`,
* for blocks use now `<% block %>` instead of `{% block %}`,
* variables can have a prefix, so the conflicting names can be rendered correctly.

Jinja cli is a copy of https://github.com/mattrobenolt/jinja2-cli (commit de5e8bf5132c80a8bbf37d788f4fff4af631753a)
Docker part is a copy of https://github.com/dinuta/jinja2docker (commit 9a44ceecd83cbe195d2d2c47e969dbb5cb5dbaa2)


## Limitations ##

### Multiple rendering ###

Missing variables are not causing errors and template can be rendered several times.
This has one limitation related to variables which have `default` filter.
All variables with a default filter will be rendered during first rendering.
The possible workaround for such variables is escaping them.

As a potential solution we may create a custom `default` filter which will be aware of prefixes.

### Prefixing ###

Let's consider such template

```yaml
<@ postgresql.user @>
<@ postgresql.db @>
```

and variables file:

```yaml
user: postgres
```

If rendering will be started with prefix `postgresql` this will be the output:

```yaml
postgres
<@ db @>
```

Prefix for db was removed. This is Jinja limitation. It shouldn't be a big problem as long
as there is no need to render the template twice with the same prefix.


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
