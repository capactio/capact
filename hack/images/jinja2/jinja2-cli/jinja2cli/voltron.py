import string
import random

from jinja2.runtime import Undefined, missing


class UndefinedDict:
    def __init__(self, parent, data):
        self.parent = parent
        self.data = data

    def __getattr__(self, attr):
        val = self.data.get(attr, None)
        if val is not None:
            return UndefinedDict(".".join([self.parent, attr]), val)
        else:
            raise AttributeError()
            return UndefinedDict(".".join([self.parent, attr]), None)

    def __str__(self):
        return str(self.data)


class Undefined(Undefined):
    __slots__ = ()

    def __str__(self):
        message = self._undefined_name
        return f"<@ {message} @>"

    def __getattr__(self, attr):
        return Undefined(name=".".join([self._undefined_name, attr]))


def random_string(letters: str = "", length: int = 10) -> str:
    """
    random_string generates random string of the given length
    using `letters`
    If no letters were provided it is using all printable letters
    except whitespaces and quotes.
    """
    if len(letters) == 0:
        # all printable except whitespaces and quotes
        printable = set(string.printable)
        whitespace = set(string.whitespace)
        q = set("\"'`")
        letters = list(printable - whitespace - q)
    return "".join(random.choices(letters, k=length))


def random_word(length: int = 10) -> str:
    """
    random word generates random word of the given length using only
    lowercase asci letters.
    """
    return random_string(letters=string.ascii_lowercase, length=length)


ALL = [random_string, random_word]
