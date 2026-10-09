from setuptools import setup, find_packages

setup(
    name="trueone-sdk",
    version="1.2.0",
    description="TrueOne Test as Code SDK for Pytest and Python Testing Projects",
    author="TrueOne Team",
    packages=find_packages(),
    python_requires=">=3.8",
    install_requires=[
        "pyyaml>=5.3",
    ],
    entry_points={
        "pytest11": [
            "trueone = aegis.plugin",
            "aegis = aegis.plugin",
        ],
    },
)
