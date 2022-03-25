import os
import string
import typing
from dataclasses import dataclass

from jinja2cli import cli, capact

import pytest

# change dir to tests directory to make relative paths possible
os.chdir(os.path.dirname(os.path.realpath(__file__)))


@dataclass
class RenderTestCase:
    name: str
    template: str
    data: typing.Dict[str, typing.Any]
    result: str

@dataclass
class PreprocessingDataTestCase:
    name: str
    config: typing.Dict[str, typing.Any]
    data: typing.Dict[str, typing.Any]
    result: typing.Dict[str, typing.Any]

render_testcases = [
    RenderTestCase(name="empty", template="", data={}, result=""),
    RenderTestCase(
        name="simple",
        template="<@ title @>",
        data={"title": b"\xc3\xb8".decode("utf8")},
        result=b"\xc3\xb8".decode("utf8"),
    ),
    RenderTestCase(
        name="prefix",
        template="<@ input.key @>",
        data={"input": {"key": "value"}},
        result="value",
    ),
    RenderTestCase(
        name="two prefixes but one provided",
        template="<@ input.key @>/<@ additionalinput.key @>",
        data={"input": {"key": "value"}},
        result="value/<@ additionalinput.key @>",
    ),
    RenderTestCase(
        name="missing prefix",
        template="<@ input.key @>",
        data={},
        result="<@ input.key @>",
    ),
    RenderTestCase(
        name="items before attrs",
        template="<@ input.values.key @>",
        data={"input": {"values": {"key": "value"}}},
        result="value",
    ),
    RenderTestCase(
        name="attrs still working",
        template="<@ input.values() @>",
        data={"input": {}},
        result="dict_values([])",
    ),
    RenderTestCase(
        name="key with dot",
        template="<@ input['foo.bar'] @>",
        data={"input": {"foo.bar": "value"}},
        result="value",
    ),
    RenderTestCase(
        name="missing key with dot",
        template='<@ input["foo.bar"] @>',
        data={},
        result='<@ input["foo.bar"] @>',
    ),
    RenderTestCase(
        name="use default value",
        template='<@ input["foo.bar"] | default("hello") @>',
        data={},
        result="hello",
    ),
    RenderTestCase(
        name="multiple dotted values",
        template='<@ input.key.key["foo.bar/baz"] | default("hello") @>',
        data={},
        result="hello",
    ),
    RenderTestCase(
        name="multiline strings",
        template="""<@ input.key.key["foo.bar/baz"] | default('hello
hello') @>""",
        data={},
        result="""hello
hello""",
    ),
]


@pytest.mark.parametrize("case", render_testcases)
def test_render(tmp_path, case):
    render_path = tmp_path / case.name
    render_path.write_text(case.template)
    output = cli.render(render_path, case.data, [])
    assert output == case.result


preprocessing_data_testcases = [
    PreprocessingDataTestCase(
        name="set prefix in the config should prefix the data",
        config={"prefix": "testprefix"},
        data = {"test": "test"},
        result={"testprefix": {"test": "test"}}
    ),
    PreprocessingDataTestCase(
        name="set unpackValue in the config should remove the value prefix",
        config={"unpackValue": True},
        data = {"value": {"test": "test"}},
        result={"test": "test"}
    ),
    PreprocessingDataTestCase(
        name="set unpackValue and prefix should output correct results",
        config={"prefix": "testprefix", "unpackValue": True},
        data = {"value": {"test": "test"}},
        result={"testprefix": {"test": "test"}}
    )
]

@pytest.mark.parametrize("case", preprocessing_data_testcases)
def test_preprocessing_data(case):
    output = cli.preprocessing_data(case.config,case.data)
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


def test_list_repr():
    assert repr(capact.List(["test"])) == '["test"]'


def contains_character_from(string, charlist):
    return True in [c in charlist for c in string]
