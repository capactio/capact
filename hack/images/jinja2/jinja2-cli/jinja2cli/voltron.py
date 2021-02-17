import string
import random

from jinja2.runtime import Undefined, missing


class VoltronUndefined(Undefined):
    __slots__ = ()

    def __str__(self):
        if self._undefined_obj is missing:
            message = self._undefined_name

        else:
            message = (
                f"no such element: {object_type_repr(self._undefined_obj)}"
                f"[{self._undefined_name!r}]"
            )

        return f"<@ {message} @>"

    def __getattr__(self, attr):
        return VoltronUndefined(name=".".join([self._undefined_name, attr]))


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
