"""
Check provides a plugin that allows you to check models and solutions.

Checking a model or a solution checks the unplanned plan units. It checks each
individual plan unit if it can be added to the solution. If the plan unit can
be added to the solution, the report will include on how many vehicles and
what the impact would be on the objective value. If the plan unit cannot be
added to the solution, the report will include the reason why it cannot be
added to the solution.

The check can be invoked on a nextroute.Model or a nextroute.Solution. If the
check is invoked on a model, an empty solution is created and the check is
executed on this empty solution. An empty solution is a solution with all the
initial stops that are fixed, initial stops that are not fixed are not added
to the solution. The check is executed on the unplanned plan units of the
solution. If the check is invoked on a solution, it is executed on the
unplanned plan units of the solution.
"""

from .schema import Objective as Objective
from .schema import ObjectiveTerm as ObjectiveTerm
from .schema import Output as Output
from .schema import PlanUnit as PlanUnit
from .schema import Solution as Solution
from .schema import Summary as Summary
from .schema import Vehicle as Vehicle
