# © 2019-present nextmv.io inc

"""
Schema for statistics.
"""

from typing import Any

from pydantic import Field

from nextroute.base_model import BaseModel


class RunStatistics(BaseModel):
    """
    Statistics about a general run.

    Parameters
    ----------
    duration : float, optional
        Duration of the run in seconds.
    iterations : int, optional
        Number of iterations.
    custom : Union[Any, Dict[str, Any]], optional
        Custom statistics created by the user. Can normally expect a `Dict[str,
        Any]`.
    """

    duration: float | None = None
    """Duration of the run in seconds."""
    iterations: int | None = None
    """Number of iterations."""
    custom: Any | dict[str, Any] | None = None
    """Custom statistics created by the user. Can normally expect a `Dict[str,
    Any]`."""


class ResultStatistics(BaseModel):
    """
    Statistics about a specific result.

    Parameters
    ----------
    duration : float, optional
        Duration of the run in seconds.
    value : float, optional
        Value of the result.
    custom : Union[Any, Dict[str, Any]], optional
        Custom statistics created by the user. Can normally expect a `Dict[str,
        Any]`.
    """

    duration: float | None = None
    """Duration of the run in seconds."""
    value: float | None = None
    """Value of the result."""
    custom: Any | dict[str, Any] | None = None
    """Custom statistics created by the user. Can normally expect a `Dict[str,
    Any]`."""


class DataPoint(BaseModel):
    """
    A data point.

    Parameters
    ----------
    x : float
        X coordinate of the data point.
    y : float
        Y coordinate of the data point.
    """

    x: float
    """X coordinate of the data point."""
    y: float
    """Y coordinate of the data point."""


class Series(BaseModel):
    """
    A series of data points.

    Parameters
    ----------
    name : str, optional
        Name of the series.
    data_points : List[DataPoint], optional
        Data of the series.
    """

    name: str | None = None
    """Name of the series."""
    data_points: list[DataPoint] | None = None
    """Data of the series."""


class SeriesData(BaseModel):
    """
    Data of a series.

    Parameters
    ----------
    value : Series, optional
        A series for the value of the solution.
    custom : List[Series], optional
        A list of series for custom statistics.
    """

    value: Series | None = None
    """A series for the value of the solution."""
    custom: list[Series] | None = None
    """A list of series for custom statistics."""


class Statistics(BaseModel):
    """
    Statistics of a solution.

    Parameters
    ----------
    run : RunStatistics, optional
        Statistics about the run.
    result : ResultStatistics, optional
        Statistics about the last result.
    series_data : SeriesData, optional
        Series data about some metric.
    statistics_schema : str, optional
        Schema (version). This class only supports `v1`.
    """

    run: RunStatistics | None = None
    """Statistics about the run."""
    result: ResultStatistics | None = None
    """Statistics about the last result."""
    series_data: SeriesData | None = None
    """Data of the series."""
    statistics_schema: str | None = Field(alias="schema", default="v1")
    """Schema (version). This class only supports `v1`."""
