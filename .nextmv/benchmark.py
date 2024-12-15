# Description:
# This script does the following:
# - Make sure the working directory is clean.
# - Pushes a new version of the app (if it does not already exist; uses git sha as version).
# - Updates the candidate instance to use the new version.
# - Runs an acceptance test between the candidate and baseline instances.
# - Waits for the test to complete.
# - Posts the result to Slack (if requested).

import os
from datetime import datetime, timezone

from nextmv import cloud

APP_ID = "nextroute-bench"
API_KEY = os.environ["NEXTMV_API_KEY"]
TAG = os.getenv("TAG", "untagged")

METRICS = [
    cloud.Metric(
        field="result.value",
        metric_type=cloud.MetricType.direct_comparison,
        params=cloud.MetricParams(
            tolerance=cloud.MetricTolerance(
                value=0.05,
                type=cloud.ToleranceType.relative,
            ),
            operator=cloud.Comparison.less_than_or_equal_to,
        ),
        statistic=cloud.StatisticType.mean,
    )
]


def check_clean_working_directory():
    if os.system("git diff --quiet") != 0 or os.system("git diff --cached --quiet") != 0:
        raise Exception("Working directory is not clean")


def run_acceptance_test() -> cloud.AcceptanceTest:
    client = cloud.Client(api_key=os.getenv("NEXTMV_API_KEY"))
    app = cloud.Application(client=client, id=APP_ID)
    ts = (
        datetime.now(timezone.utc)
        .replace(microsecond=0)
        .isoformat()
        .replace("+00:00", "Z")
        .replace(":", "")
        .replace("-", "")
    )
    id = f"auto-{TAG}-{ts}"
    print(f"Running acceptance test with ID: {id}")
    print("Waiting for the test to complete...")
    result = app.new_acceptance_test_with_result(
        candidate_instance_id="candidate",
        baseline_instance_id="baseline",
        id=id,
        metrics=METRICS,
        name=f"Auto-test {TAG}",
        description=f"Automated test for {TAG}",
        input_set_id="nextroute-bench-v20",
        polling_options=cloud.PollingOptions(
            max_duration=600,  # 10 minutes
            max_tries=1000,  # basically forever - we'll stop by duration
        ),
    )
    passed = "passed" if result and result.results and result.results.passed else "unknown"
    print(f"Acceptance test completed with status: {passed}")
    return result


def main():
    run_acceptance_test()


if __name__ == "__main__":
    main()
