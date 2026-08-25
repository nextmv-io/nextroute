# © 2019-present nextmv.io inc

"""
JSON class for data wrangling JSON objects.
"""

from typing import Any

from pydantic import BaseModel


class BaseModel(BaseModel):
    """Base class for data wrangling tasks with JSON."""

    @classmethod
    def from_dict(cls, data: dict[str, Any]):
        """Instantiates the class from a dict."""

        return cls(**data)

    def to_dict(self) -> dict[str, Any]:
        """Converts the class to a dict."""

        return self.model_dump(mode="json", exclude_none=True, by_alias=True)
