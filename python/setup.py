from setuptools import setup, find_packages

VERSION = '0.0.1'
DESCRIPTION = 'The official Monarch C2 integration package'

setup(
    name="monarch",
    version=VERSION,
    author="Pygrum",
    url="https://github.com/pygrum/monarch",
    description=DESCRIPTION,
    license="MIT",
    packages=find_packages(),
    requires=[],
)
