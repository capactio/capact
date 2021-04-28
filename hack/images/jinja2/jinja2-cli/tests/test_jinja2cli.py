import os
import string

from jinja2cli import cli, capact

# change dir to tests directory to make relative paths possible
os.chdir(os.path.dirname(os.path.realpath(__file__)))


def test_relative_path():
    path = "./files/template.j2"

    title = b"\xc3\xb8".decode("utf8")
    output = cli.render(path, {"title": title}, [])
    assert output == title
    assert type(output) == cli.text_type


def test_absolute_path():
    absolute_base_path = os.path.dirname(os.path.realpath(__file__))
    path = os.path.join(absolute_base_path, "files", "template.j2")

    title = b"\xc3\xb8".decode("utf8")
    output = cli.render(path, {"title": title}, [])
    assert output == title
    assert type(output) == cli.text_type


def test_random_password():
    absolute_base_path = os.path.dirname(os.path.realpath(__file__))
    path = os.path.join(absolute_base_path, "files", "random_password.j2")

    output = cli.render(path, {}, [])
    assert contains_character_from(output, string.ascii_lowercase)
    assert contains_character_from(output, string.ascii_uppercase)
    assert contains_character_from(output, string.digits)
    assert contains_character_from(output, capact._punctuation)
    assert type(output) == cli.text_type


def contains_character_from(string, charlist):
    return True in [c in charlist for c in string]
