import string
import random
import base64

from jinja2.runtime import Undefined, missing

_punctuation = r"""!"#$%&()*+,-./:;<=>?@[\]^_{|}~"""


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


def random_password(length: int = 10, numbers=True,
                    lowercase=True, uppercase=True, special=True) -> str:
    """
    random_password generates a random password with a given length
    and requirements.
    It ensures at least one character from the selected groups
    will be present in the generated password
    """
    chars = ""
    pwlist = []

    if numbers:
        chars += string.digits
        pwlist += [random.choice(string.digits)]
    if lowercase:
        chars += string.ascii_lowercase
        pwlist += [random.choice(string.ascii_lowercase)]
    if uppercase:
        chars += string.ascii_uppercase
        pwlist += [random.choice(string.ascii_uppercase)]
    if special:
        chars += _punctuation
        pwlist += [random.choice(_punctuation)]

    to_fill = length - numbers - lowercase - uppercase - special
    pwlist += [random.choice(chars) for i in range(to_fill)]

    random.shuffle(pwlist)
    return ''.join(pwlist)


def random_word(length: int = 10) -> str:
    """
    random word generates random word of the given length using only
    lowercase asci letters.
    """
    return random_string(letters=string.ascii_lowercase, length=length)


def b64encode(s: str) -> str:
    return base64.b64encode(s.encode()).decode()


def b64decode(s: str) -> str:
    return base64.b64decode(s).decode()


GLOBALS = [random_string, random_word, random_password]
FILTERS = [b64decode, b64encode]
