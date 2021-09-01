import os
import string
import typing
from dataclasses import dataclass

from jinja2cli import cli, capact

import pytest

# change dir to tests directory to make relative paths possible
os.chdir(os.path.dirname(os.path.realpath(__file__)))


@dataclass
class TestCase:
    name: str
    template: str
    data: typing.Dict[str, typing.Any]
    result: str


render_testcases = [
    TestCase(name="empty", template="", data={}, result=""),
    TestCase(
        name="simple",
        template="<@ title @>",
        data={"title": b"\xc3\xb8".decode("utf8")},
        result=b"\xc3\xb8".decode("utf8"),
    ),
    TestCase(
        name="prefix",
        template="<@ input.key @>",
        data={"input": {"key": "value"}},
        result="value",
    ),
    TestCase(
        name="two prefixes but one provided",
        template="<@ input.key @>/<@ additionalinput.key @>",
        data={"input": {"key": "value"}},
        result="value/<@ additionalinput.key @>",
    ),
    TestCase(
        name="missing prefix",
        template="<@ input.key @>",
        data={},
        result="<@ input.key @>",
    ),
    TestCase(
        name="items before attrs",
        template="<@ input.values.key @>",
        data={"input": {"values": {"key": "value"}}},
        result="value",
    ),
    TestCase(
        name="attrs still working",
        template="<@ input.values() @>",
        data={"input": {}},
        result="dict_values([])",
    ),
    TestCase(
        name="key with dot",
        template="<@ input['foo.bar'] @>",
        data={"input": {"foo.bar": "value"}},
        result="value",
    ),
    TestCase(
        name="missing key with dot",
        template='<@ input["foo.bar"] @>',
        data={},
        result='<@ input["foo.bar"] @>',
    ),
    TestCase(
        name="use default value",
        template='<@ input["foo.bar"] | default("hello") @>',
        data={},
        result="hello",
    ),
]


@pytest.mark.parametrize("case", render_testcases)
def test_render(tmp_path, case):
    render_path = tmp_path / case.name
    render_path.write_text(case.template)
    output = cli.render(render_path, case.data, [])
    assert output == case.result


def test_random_password(tmp_path):
    random_pass_path = tmp_path / "random.template"
    random_pass_path.write_text("<@ random_password(length=4) @>")

    output = cli.render(random_pass_path, {}, [])
    assert contains_character_from(output, string.ascii_lowercase)
    assert contains_character_from(output, string.ascii_uppercase)
    assert contains_character_from(output, string.digits)
    assert contains_character_from(output, capact._punctuation)
    assert type(output) == cli.text_type


def contains_character_from(string, charlist):
    return True in [c in charlist for c in string]
