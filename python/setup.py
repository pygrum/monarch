from setuptools import setup, find_packages

VERSION = '0.0.3'
DESCRIPTION = 'The official Monarch C2 integration package'

setup(
    name="monarch_c2_sdk",
    version=VERSION,
    author="Pygrum",
    url="https://github.com/pygrum/monarch",
    description=DESCRIPTION,
    license="MIT",
    packages=find_packages(),
    requires=[],
)
